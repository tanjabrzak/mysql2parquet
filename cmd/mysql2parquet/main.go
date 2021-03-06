package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/in4it/mysql2parquet/pkg/mysql"
	"github.com/in4it/mysql2parquet/pkg/parquet"
)

func main() {
	var (
		connectionString string
		query            string
		out              string
		compression      string
		debug            string
	)

	flag.StringVar(&connectionString, "connectionString", "", "MySQL connectionstring")
	flag.StringVar(&query, "query", "", "query")
	flag.StringVar(&out, "out", "", "outputfile")
	flag.StringVar(&compression, "compression", "none", "compression to apply (snappy/bzip/gzip)")
	flag.StringVar(&debug, "debug", "no", "enable debug")

	flag.Parse()

	// do mysql query
	m := mysql.New()
	m.Init(connectionString)
	m.Query(query)

	// initialize parquet file and schema
	columnNames, columnTypes := m.GetColumnInfo()
	schema := getSchema(columnNames, columnTypes)
	p := parquet.NewWriter()
	p.Open(out, schema, compression)

	if debug == "true" {
		fmt.Printf("Schema: %+v", schema)
	}

	for {
		row, nextRow := m.GetRow()
		if !nextRow {
			break
		}
		data := make([]string, len(columnNames))
		for k, v := range row {
			data[k] = toParquetValue(v.RowData, v.RowType)
		}
		if debug == "true" {
			fmt.Printf("Data: %+v\n", data)
		}
		p.WriteLine(data)
	}

	m.RowClose()
	m.Close()
	p.Close()

}

func toParquetValue(value []byte, rowType string) string {
	var ret string
	switch rowType {
	case "DATE":
		layout := "2006-01-02"
		t, err := time.Parse(layout, string(value))
		if err != nil {
			panic(fmt.Errorf("Couldn't convert DATE value (%s): %s", string(value), err))
		}
		ret = strconv.FormatInt(t.Unix(), 10)
	case "DATETIME":
		layout := "2006-01-02 15:04:05"
		t, err := time.Parse(layout, string(value))
		if err != nil {
			panic(fmt.Errorf("Couldn't convert DATETIME value (%s): %s", string(value), err))
		}
		ret = strconv.FormatInt(t.Unix(), 10)
	case "TIMESTAMP":
		layout := "2006-01-02 15:04:05"
		t, err := time.Parse(layout, string(value))
		if err != nil {
			panic(fmt.Errorf("Couldn't convert TIMESTAMP value (%s): %s", string(value), err))
		}
		ret = strconv.FormatInt(t.Unix(), 10)
	default:
		ret = string(value)
	}
	return ret
}

func MySQLToParquetType(mysqlType string) string {
	parquetType := ""
	switch mysqlType {
	case "BOOLEAN":
		parquetType = "BOOL"
	case "FLOAT":
		parquetType = "FLOAT"
	case "DOUBLE":
		parquetType = "DOUBLE"
	case "DECIMAL":
		parquetType = "DOUBLE"
	case "VARCHAR":
		parquetType = "UTF8"
	case "CHAR":
		parquetType = "UTF8"
	case "TINYTEXT":
		parquetType = "UTF8"
	case "TEXT":
		parquetType = "UTF8"
	case "BLOB":
		parquetType = "BYTE_ARRAY"
	case "MEDIUMTEXT":
		parquetType = "UTF8"
	case "MEDIUMBLOB":
		parquetType = "BYTE_ARRAY"
	case "LONGTEXT":
		parquetType = "UTF8"
	case "LONGBLOB":
		parquetType = "BYTE_ARRAY"
	case "TINYINT":
		parquetType = "INT32"
	case "SMALLINT":
		parquetType = "INT32"
	case "MEDIUMINT":
		parquetType = "INT32"
	case "INT":
		parquetType = "INT32"
	case "BIGINT":
		parquetType = "INT64"
	case "DATE":
		parquetType = "DATE"
	case "DATETIME":
		parquetType = "TIMESTAMP_MILLIS"
	case "TIMESTAMP":
		parquetType = "TIMESTAMP_MILLIS"

	default:
		panic(fmt.Errorf("Encoding not found: %s", mysqlType))
	}
	return parquetType
}

func getSchema(columnNames, columnTypes []string) []string {
	ret := []string{}
	for k, v := range columnNames {
		parquetType := MySQLToParquetType(columnTypes[k])
		if parquetType == "UTF8" {
			ret = append(ret, fmt.Sprintf("name=%s, type=%s, encoding=PLAIN_DICTIONARY", v, parquetType))
		} else {
			ret = append(ret, fmt.Sprintf("name=%s, type=%s", v, parquetType))
		}
	}
	return ret
}
