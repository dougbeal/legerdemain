// +build ignore

package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/shurcooL/vfsgen"
	"github.com/shurcooL/httpfs/union"
	
)

func main() {
	var cwd, _ = os.Getwd()
	templates := http.Dir(filepath.Join(cwd, "assets/templates"))
	static := http.Dir(filepath.Join(cwd, "assets/static"))
	// Assets contains the project's assets.
	var assets http.FileSystem = union.New(map[string]http.FileSystem{
		"/templates": templates,
		"/static":    static,
	})

	if err := vfsgen.Generate(assets, vfsgen.Options{
		PackageName:  "plaid",
		VariableName: "assets",
	}); err != nil {
		log.Fatalln(err)
	}
}
