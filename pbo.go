package pbo

import (
	"fmt"
	"io"
)

type PBO struct {
	Properties map[string]string
	Files      []File

	dataOffset uint32
}

func Load(in io.ReadSeeker) (*PBO, error) {
	bank := PBO{
		Files: []File{},
	}

	for {
		file, err := readHeader(in)
		if err != nil {
			return nil, err
		}

		bank.dataOffset += uint32(len(file.Filename)) + 21

		// first header (optional) marks properties list
		if file.Type == Vers {
			properties, offset, err := readProperties(in)
			if err != nil {
				return nil, err
			}
			bank.Properties = properties
			bank.dataOffset += offset
			continue
		}

		// final header, marks end of file list
		if len(file.Filename) == 0 {
			break
		}

		bank.Files = append(bank.Files, *file)
	}

	// update read offset for each header
	cumulativeOffset := uint32(0)
	for index, file := range bank.Files {
		file.offset = bank.dataOffset + cumulativeOffset
		file.reader = in
		bank.Files[index] = file
		cumulativeOffset += file.DataSize
	}

	return &bank, nil
}

func readProperties(in io.Reader) (map[string]string, uint32, error) {
	var offset uint32 = 0
	propertyStrings := []string{}
	for {
		str, err := readAsciiz(in)
		if err != nil {
			return nil, offset, err
		}
		offset += uint32(len(str)) + 1
		// empty string marks end of properties list
		if str == "" {
			break
		}
		propertyStrings = append(propertyStrings, str)
	}

	if len(propertyStrings)%2 != 0 {
		return nil, offset, fmt.Errorf("invalid properties list, expected even number of strings, got %d", len(propertyStrings))
	}

	properties := make(map[string]string, len(propertyStrings)/2)
	for i := 0; i < len(propertyStrings); i += 2 {
		properties[propertyStrings[i]] = propertyStrings[i+1]
	}

	return properties, offset, nil
}
