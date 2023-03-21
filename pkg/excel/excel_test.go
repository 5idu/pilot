package excel

import (
	"testing"
)

func TestExportExcel(t *testing.T) {
	headers := []string{"用户名", "性别", "年龄"}
	values := [][]interface{}{
		{"bob", "1", "男"},
		{"kitty", "1", "女"},
	}
	f, err := ExportExcel("Sheet1", headers, values)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.SaveAs("test.xlsx"); err != nil {
		t.Fatal(err)
	}
	t.Log("done.")
}
