package sumfile

import (
	"bytes"
	"maps"
	"os"
	"path/filepath"
	"slices"
)

const sumFilename = "gengo.sum"

func Load(modRoot string) (*File, error) {
	data, err := os.ReadFile(filepath.Join(modRoot, sumFilename))
	if err != nil {
		return nil, err
	}

	sum := &File{
		Dir:  modRoot,
		Data: map[string]string{},
	}

	for line := range bytes.Lines(data) {
		parts := bytes.Fields(line)
		if len(parts) >= 2 {
			sum.Data[string(parts[0])] = string(parts[1])
		}
	}

	return sum, nil
}

type File struct {
	Dir  string
	Data map[string]string
}

func (f *File) Save() error {
	file, err := os.OpenFile(filepath.Join(f.Dir, sumFilename), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(f.Bytes())
	return err
}

func (f *File) Bytes() []byte {
	b := bytes.NewBuffer(nil)
	for _, pkgPath := range slices.Sorted(maps.Keys(f.Data)) {
		b.WriteString(pkgPath)
		b.WriteString(" ")
		b.WriteString(f.Data[pkgPath])
		b.WriteString("\n")
	}
	return b.Bytes()
}

func (f *File) Sum(pkgPath string) string {
	if f.Data == nil {
		return ""
	}
	return f.Data[pkgPath]
}
