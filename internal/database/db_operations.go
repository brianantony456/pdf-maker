package database

import (
	"fmt"
	"strconv"
	"time"
)

func (s *serviceImpl) AddColumn(name, dataType string) error {
	query := fmt.Sprintf(`ALTER TABLE data_entries ADD COLUMN %s %s`, name, dataType)
	_, err := s.db.Exec(query)
	return err
}

func (s *serviceImpl) RenameColumn(oldName, newName string) error {
	tempTable := "temp_" + oldName
	columns, err := s.GetColumns()
	if err != nil {
		return err
	}

	// Create a temporary table with the new schema
	createQuery := "CREATE TABLE " + tempTable + " ("
	for _, column := range columns {
		if column["name"] == oldName {
			createQuery += fmt.Sprintf("%s %s,", newName, column["data_type"])
		} else {
			createQuery += fmt.Sprintf("%s %s,", column["name"], column["data_type"])
		}
	}
	createQuery = createQuery[:len(createQuery)-1] + ");"
	_, err = s.db.Exec(createQuery)
	if err != nil {
		return fmt.Errorf("failed to create temporary table: %v", err)
	}

	// Copy data from original table to temp table
	insertQuery := fmt.Sprintf(`INSERT INTO %s SELECT `, tempTable)
	for _, column := range columns {
		if column["name"] == oldName {
			insertQuery += fmt.Sprintf("%s as %s,", oldName, newName)
		} else {
			insertQuery += fmt.Sprintf("%s,", column["name"])
		}
	}
	insertQuery = insertQuery[:len(insertQuery)-1] + " FROM data_entries;"
	_, err = s.db.Exec(insertQuery)
	if err != nil {
		return fmt.Errorf("failed to copy data to temporary table: %v", err)
	}

	// Drop original table and rename temp table
	_, err = s.db.Exec("DROP TABLE data_entries;")
	if err != nil {
		return fmt.Errorf("failed to drop original table: %v", err)
	}
	_, err = s.db.Exec(fmt.Sprintf("ALTER TABLE %s RENAME TO data_entries;", tempTable))
	if err != nil {
		return fmt.Errorf("failed to rename temp table: %v", err)
	}

	return nil
}

func (s *serviceImpl) RemoveColumn(name string) error {
	tempTable := "temp_" + name
	columns, err := s.GetColumns()
	if err != nil {
		return err
	}

	// Create a temporary table without the column to be removed
	createQuery := "CREATE TABLE " + tempTable + " ("
	for _, column := range columns {
		if column["name"] != name {
			createQuery += fmt.Sprintf("%s %s,", column["name"], column["data_type"])
		}
	}
	createQuery = createQuery[:len(createQuery)-1] + ");"
	_, err = s.db.Exec(createQuery)
	if err != nil {
		return fmt.Errorf("failed to create temp table: %v", err)
	}

	// Copy data to temp table, excluding the column to be removed
	insertQuery := fmt.Sprintf(`INSERT INTO %s SELECT `, tempTable)
	for _, column := range columns {
		if column["name"] != name {
			insertQuery += fmt.Sprintf("%s,", column["name"])
		}
	}
	insertQuery = insertQuery[:len(insertQuery)-1] + " FROM data_entries;"
	_, err = s.db.Exec(insertQuery)
	if err != nil {
		return fmt.Errorf("failed to copy data to temp table: %v", err)
	}

	// Drop original table and rename temp table
	_, err = s.db.Exec("DROP TABLE data_entries;")
	if err != nil {
		return fmt.Errorf("failed to drop original table: %v", err)
	}
	_, err = s.db.Exec(fmt.Sprintf("ALTER TABLE %s RENAME TO data_entries;", tempTable))
	if err != nil {
		return fmt.Errorf("failed to rename temp table: %v", err)
	}

	return nil
}

func (s *serviceImpl) AddRow(data map[string]string) error {
	insertQuery := "INSERT INTO data_entries ("
	valuesQuery := "VALUES ("
	args := []interface{}{}

	for key, value := range data {
		insertQuery += fmt.Sprintf("%s,", key)
		valuesQuery += "?,"
		args = append(args, value)
	}

	insertQuery = insertQuery[:len(insertQuery)-1] + ") "
	valuesQuery = valuesQuery[:len(valuesQuery)-1] + ")"
	query := insertQuery + valuesQuery

	_, err := s.db.Exec(query, args...)
	return err
}

func (s *serviceImpl) GetRow(id int) (map[string]string, error) {
	columns, err := s.GetColumns()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, "
	for _, column := range columns {
		query += fmt.Sprintf("%s, ", column["name"])
	}
	query = query[:len(query)-2] + " FROM data_entries WHERE id = ?"

	row := s.db.QueryRow(query, id)
	result := make(map[string]string)
	args := make([]interface{}, len(columns)+1)
	for i := range args {
		args[i] = new(string)
	}

	err = row.Scan(args...)
	if err != nil {
		return nil, err
	}

	for i, column := range columns {
		result[column["name"].(string)] = *(args[i+1].(*string))
	}
	result["id"] = strconv.Itoa(id)
	return result, nil
}

func (s *serviceImpl) UpdateRow(id int, data map[string]string) error {
	query := "UPDATE data_entries SET "
	args := []interface{}{}

	for key, value := range data {
		query += fmt.Sprintf("%s = ?, ", key)
		args = append(args, value)
	}

	query = query[:len(query)-2] + " WHERE id = ?"
	args = append(args, id)

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return err
	}

	return s.logUpdate(id, data)
}

func (s *serviceImpl) GetColumns() ([]map[string]interface{}, error) {
	rows, err := s.db.Query(`PRAGMA table_info(data_entries);`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []map[string]interface{}
	for rows.Next() {
		var cid int
		var name, dataType, notnull, dfltValue, pk interface{}
		err := rows.Scan(&cid, &name, &dataType, &notnull, &dfltValue, &pk)
		if err != nil {
			return nil, err
		}
		columns = append(columns, map[string]interface{}{
			"name": name, "data_type": dataType, "notnull": notnull, "dflt_value": dfltValue, "pk": pk,
		})
	}
	return columns, nil
}

func (s *serviceImpl) logUpdate(id int, data map[string]string) error {
	timestamp := time.Now().Format(time.RFC3339)
	dataStr := fmt.Sprintf("%v", data)

	query := `INSERT INTO update_log (id, timestamp, data) VALUES (?, ?, ?)`
	_, err := s.db.Exec(query, id, timestamp, dataStr)
	return err
}

func (s *serviceImpl) GetAllRows() ([]map[string]interface{}, error) {
	columns, err := s.GetColumns()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, "
	for _, column := range columns {
		query += fmt.Sprintf("%s, ", column["name"])
	}
	query = query[:len(query)-2] + " FROM data_entries"

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		row := make([]interface{}, len(columns)+1)
		rowPtrs := make([]interface{}, len(columns)+1)
		for i := range row {
			rowPtrs[i] = &row[i]
		}

		err = rows.Scan(rowPtrs...)
		if err != nil {
			return nil, err
		}

		rowData := make(map[string]interface{})
		rowData["id"] = row[0]
		for i, col := range columns {
			rowData[col["name"].(string)] = row[i+1]
		}
		result = append(result, rowData)
	}
	return result, nil
}
