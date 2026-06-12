package output

import (
	"fmt"
	"os"
)

func WriteSQL(sqlText, outputFile string) error {
	if outputFile == "" {
		_, err := fmt.Print(sqlText)
		return err
	}
	return os.WriteFile(outputFile, []byte(sqlText), 0o644)
}
