package headers

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type Headers map[string]string
func NewHeaders() Headers{
	return make(map[string]string)
}


func (header Headers) Parse(data []byte) (n int, done bool, err error){
	
	var idx int = bytes.Index(data, []byte("\r\n"))
	if idx == -1{
		return 0, false, nil
	}
	//end of headers
	if idx == 0{
		return 2, true, nil
	}
	var line string = string(data[:idx])
	var parts []string = strings.SplitN(line, ":", 2)
	if len(parts) < 2{
		return 0, false, errors.New("Invalid header line: " + line)
	}
	var key string = strings.TrimLeft(parts[0], " ")
	if strings.HasSuffix(key, " ") {
		return 0, false, errors.New("Header key cannot have trailing spaces: " + line)
	}

	var value string = strings.TrimSpace(parts[1])
	if key == "" || value == ""{
		return 0, false, errors.New("Header key and value cannot be empty: " + line)
	}

	if !isValidTokenStr(key){
		return 0, false, errors.New("Header key contains invalid characters: " + line)
	}

	key = strings.ToLower(key)

	if _, exists := header[key]; exists{
		value = fmt.Sprintf("%s, %s", header[key], value)
	}

	header[key] = value
	return idx + 2, false, nil

}

func isValidTokenStr(str string) bool{
	for _, char := range str {
		if (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			strings.ContainsRune("!#$%&'*+-.^_`|~", char) {
				continue
		}else{
			return false
		}
	}
	return true
}

func (header Headers) Get(key string) (string, bool){
	key = strings.ToLower(key)
	value, exists := header[key]
	return value, exists
}

func (header *Headers) Override(key string, value string){
	key = strings.ToLower(key)
	(*header)[key] = value
}

func (header *Headers) Add(key string, values ...string){
	key = strings.ToLower(key)
	combined := strings.Join(values, ", ")
	existing, exists := (*header)[key]
	if exists && existing != "" {
		(*header)[key] = fmt.Sprintf("%s, %s", existing, combined)
	}else{
		(*header)[key] = combined
	}
}

func (header *Headers) Remove(key string){
	key = strings.ToLower(key)
	delete(*header, key)
}
