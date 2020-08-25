package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

func main() {
	path := "/home/andres/Desktop/pruebaCreateTable.sql"
	reader := DBFileReaderImpl{}
	fields, tablename, domain := reader.Read(path)
	fmt.Println(fields)
	fmt.Println(tablename)
	fmt.Println(domain)
	fmt.Println(createModel(fields, domain))
	fmt.Println(createMapper(fields, domain, tablename))
}

type dBFileReader interface {
	Read(path string) []DbField
}

type DBFileReaderImpl struct {
}

func (DBFileReaderImpl) Read(path string) ([]DbField, string, string) {
	dat, err := ioutil.ReadFile(path)
	check(err)
	dataString := strings.ToLower(string(dat))
	r, _ := regexp.Compile("\"([a-z_]+[a-z]+)\"")
	fmt.Println(dataString)
	//fmt.Println(r.MatchString(dataString))
	tableName := r.FindString(dataString)
	fmt.Println(tableName)

	start := strings.Index(dataString, "(") + 1
	end := strings.LastIndex(dataString, ")")
	fieldsString := strings.TrimSpace(dataString[start:end])
	fieldsString = strings.Join(strings.Fields(strings.TrimSpace(fieldsString)), " ")

	reg := regexp.MustCompile(`\([^"]*\)`)
	fieldsString = reg.ReplaceAllString(fieldsString, "${1}")
	fieldsString = strings.ReplaceAll(fieldsString, "\"", "")
	fieldsSlice := strings.Split(fieldsString, ",")
	fmt.Println(fieldsSlice)
	dbfields := []DbField{}
	for _, field := range fieldsSlice {
		field = strings.TrimSpace(field)
		dbfields = append(dbfields, generateField(field))
	}
	tableName = strings.ReplaceAll(tableName, "\"", "")
	domainName := normalizeName(tableName)
	domainName = strings.ToUpper(string(domainName[0])) + domainName[1:]
	return dbfields, tableName, domainName
}

type DbField struct {
	JavaName string
	DbName   string
	JavaType string
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func generateField(in string) DbField {
	javaTypeMap := map[string]string{
		"varchar2": "String",
		"number":   "Integer",
		"date":     "Date",
	}

	values := strings.Split(in, " ")
	fname := values[0]
	ftype := values[1]
	fmt.Println("fname =" + fname + ",ftype =" + ftype)
	return DbField{JavaName: normalizeName(fname), DbName: fname, JavaType: javaTypeMap[ftype]}
}

func normalizeName(in string) string {
	result := in
	indexes := getIndexes(in)
	fmt.Println(indexes)
	for _, index := range indexes {
		result = replaceStringAt(result, strings.ToUpper(string(result[index+1])), index+1)
	}
	result = strings.ReplaceAll(result, "_", "")
	return result
}

func getIndexes(in string) []int {
	result := []int{}
	strToProcess := in
	for strings.Index(strToProcess, "_") != -1 {
		index := strings.Index(strToProcess, "_")
		toinsert := len(result) + index
		result = append(result, toinsert)
		strToProcess = strToProcess[0:index] + strToProcess[index+1:]
	}
	return result
}

func replaceStringAt(str string, nstring string, index int) string {
	return str[0:index] + nstring + str[index+len(nstring):]

}

func createModel(fields []DbField, modelName string) string {
	modelTemplate := `

package cl.andres.bpossnative.equipos.modelo;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
	
@Builder
@NoArgsConstructor
@AllArgsConstructor
@Data
public class %s {
	%s
	}
	
	`
	fieldTemplate := `private  %s %s;
	`
	fieldString := ""
	for _, field := range fields {
		fieldString = fieldString + fmt.Sprintf(fieldTemplate, field.JavaType, field.JavaName)
	}

	return fmt.Sprintf(modelTemplate, modelName, fieldString)
}

func createMapper(dbFields []DbField, domain string, tablename string) string {
	mapperTemplate := `package cl.andres.bpossnative.equipos.mapper;

import cl.andres.bpossnative.equipos.modelo.%[1]s;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;

import java.util.List;
@Mapper
public interface %[1]sMapper {
	%[2]s
	}`

	selectTemplate := `
	@Select("<script> select " +
            %s
        " from %s " +
        " where 1=1 " +
            %s
            "</script>"
	)
	List<%[4]s> find%[4]s(%[4]s req);
	`
	fieldsSelectTemplate := `
	"%s %s,"+`
	conditionsSelectTemplate := `"<if test ='%[1]s !=null'> and %[2]s=#{%[1]s} </if>" + 
	`
	fieldSelectString := ""
	conditionSelectString := ""

	for _, dbField := range dbFields {
		fieldSelectString = fieldSelectString + fmt.Sprintf(fieldsSelectTemplate, dbField.DbName, dbField.JavaName)
		conditionSelectString = conditionSelectString +
			fmt.Sprintf(conditionsSelectTemplate, dbField.JavaName, dbField.DbName)
	}
	indexCommaToRemove := strings.LastIndex(fieldSelectString, ",")
	fieldSelectString = replaceStringAt(fieldSelectString, " ", indexCommaToRemove)
	selectString := fmt.Sprintf(selectTemplate, fieldSelectString, tablename, conditionSelectString, domain)
	//TODO AGREGAR INSERT
	return fmt.Sprintf(mapperTemplate, domain, selectString)
}
