// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

/*
Package po provides support for reading and writing GNU PO file.

Examples:
	import (
		"github.com/chai2010/gettext-go/gettext/po"
	)

	func main() {
		poFile, err := po.Load("test.po")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%v", poFile)
	}

The GNU PO file specification is at
http://www.gnu.org/software/gettext/manual/html_node/PO-Files.html.
*/
package po
