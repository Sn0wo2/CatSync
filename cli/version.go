package cli

import (
	"fmt"
	"os"

	"github.com/Sn0wo2/CatSync/version"
)

func handleVersion() {
	fmt.Println(version.GetCLIVersion())
	os.Exit(0)
}
