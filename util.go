package pbo

import "io"

func readAsciiz(in io.Reader) (string, error) {
	out := ""
	buf := make([]byte, 1)
	for {
		_, err := in.Read(buf)
		if err != nil {
			return "", err
		}
		if buf[0] == 0 {
			break
		}
		out = out + string(buf[0]) // TODO
	}
	return out, nil
}
