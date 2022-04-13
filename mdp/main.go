package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"html/template"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

const (
	defaultTemplate = `<!DOCTYPE html>
<html>
	<head>
		<meta http-equiv="content-type" content="text/html; charset=utf-8">
		<title>{{.Title}}</title>
	</head>
	<body>
		{{.Body}}
	</body>
</html>
`
)

type content struct {
	Title string
	Body template.HTML
}

func main() {
	// parse flags
	filename := flag.String("file", "", "Markdown file to preview")
	skipPreview := flag.Bool("s", false, "Skip auto-preview")
	tFname := flag.String("t", "", "Alternate template name")
	flag.Parse()

	if *filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(*filename, *tFname, os.Stdout, *skipPreview); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(filename, tFname string, out io.Writer, skipPreview bool) error {
	// read all the data from the input file and check the error
	input, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	htmlData, err := parseContent(input, tFname)
	if err != nil {
		return err
	}

	/* basename := filepath.Base(filename)
	outName := fmt.Sprintf("%s.html", strings.TrimSuffix(basename, filepath.Ext(basename))) */

	// create temporary file and check for errors
	temp, err := os.CreateTemp("", "mdp*.html")
	if err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}

	outName := temp.Name()
	fmt.Fprintln(out, outName)

	if err := saveHTML(outName, htmlData); err != nil {
		return err
	}

	if !skipPreview {
		return preview(outName)
	}

	// clean up the file
	defer os.Remove(outName)

	return nil
}

func parseContent(input []byte, tFname string) ([]byte, error) {
	// parse the markdown file through blackfriday and bluemonday to generate a
	// valid and safe HTML
	output := blackfriday.Run(input)
	body := bluemonday.UGCPolicy().SanitizeBytes(output)

	// parse the content of defaultTemplate const into a new template instance
	t, err := template.New("mdp").Parse(defaultTemplate)

	if err != nil {
		return nil, err
	}

	// if user provided alternate template file, use it
	if tFname != "" {
		t, err = template.ParseFiles(tFname)
		if err != nil {
			return nil, err
		}
	}

	c := content {
		Title: "Markdown Preview Tool",
		Body: template.HTML(body),
	}

	var buffer bytes.Buffer

	// execute the template with content-type
	if err := t.Execute(&buffer, c); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func saveHTML(outFnName string, data []byte) error {
	return os.WriteFile(outFnName, data, 0644)
}

func preview(fname string) error {
	cName := ""
	cParams := []string{}

	// define executable based on OS
	switch runtime.GOOS {
	case "linux":
		cName = "xdg-open"
	case "windows":
		cName = "cmd.exe"
		cParams = []string{"/C", "start"}
	case "darwin":
		cName = "open"
	default:
		return fmt.Errorf("OS not supported")
	}

	// Append filename to parameters slice
	cParams = append(cParams, fname)

	// Locate executable in PATH
	cPath, err := exec.LookPath(cName)
	if err != nil {
		return err
	}

	// Open the file using default program
	err =  exec.Command(cPath, cParams...).Run()

	// add the delay, so program waits before quiting
	time.Sleep(10 * time.Second)
	// method is not reliable, because it can take much more time to open
	// browser in old systesm

	return err
}
