// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package gowsdl

import (
	"bytes"
	"go/format"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestElementGenerationDoesntCommentOutStructProperty(t *testing.T) {
	g := GoWSDL{
		file:         "fixtures/test.wsdl",
		pkg:          "myservice",
		makePublicFn: makePublic,
	}

	resp, err := g.Start()
	if err != nil {
		t.Error(err)
	}

	if strings.Contains(string(resp["types"]), "// this is a comment  GetInfoResult string `xml:\"GetInfoResult,omitempty\"`") {
		t.Error("Type comment should not comment out struct type property")
		t.Error(string(resp["types"]))
	}
}

func TestVboxGeneratesWithoutSyntaxErrors(t *testing.T) {
	files, err := filepath.Glob("fixtures/*.wsdl")
	if err != nil {
		t.Error(err)
	}

	for _, file := range files {
		g := GoWSDL{
			file:         file,
			pkg:          "myservice",
			makePublicFn: makePublic,
		}

		resp, err := g.Start()
		if err != nil {
			continue
			//t.Error(err)
		}

		data := new(bytes.Buffer)
		data.Write(resp["header"])
		data.Write(resp["types"])
		data.Write(resp["operations"])
		data.Write(resp["soap"])

		_, err = format.Source(data.Bytes())
		if err != nil {
			t.Error(err)
		}
	}
}

func TestSOAPHeaderGeneratesWithoutErrors(t *testing.T) {
	g := GoWSDL{
		file:         "fixtures/ferry.wsdl",
		pkg:          "myservice",
		makePublicFn: makePublic,
	}

	resp, err := g.Start()
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(string(resp["operations"]), "SetHeader") {
		t.Error("SetHeader method should be generated in the service operation")
	}
}

func TestEnumerationsGeneratedCorrectly(t *testing.T) {
	enumStringTest := func(t *testing.T, fixtureWsdl string, varName string, typeName string, enumString string) {
		g := GoWSDL{
			file:         "fixtures/" + fixtureWsdl,
			pkg:          "myservice",
			makePublicFn: makePublic,
		}

		resp, err := g.Start()
		if err != nil {
			t.Error(err)
		}

		re := regexp.MustCompile(varName + " " + typeName + " = \"([^\"]+)\"")
		matches := re.FindStringSubmatch(string(resp["types"]))

		if len(matches) != 2 {
			t.Errorf("No match or too many matches found for %s", varName)
		} else if matches[1] != enumString {
			t.Errorf("%s got '%s' but expected '%s'", varName, matches[1], enumString)
		}
	}
	enumStringTest(t, "chromedata.wsdl", "DriveTrainFrontWheelDrive", "DriveTrain", "Front Wheel Drive")
	enumStringTest(t, "vboxweb.wsdl", "SettingsVersionV1_14", "SettingsVersion", "v1_14")

}

func TestSimpleTypeList(t *testing.T) {
	g := GoWSDL{
		file:         "fixtures/list.xsd",
		pkg:          "myservice",
		makePublicFn: makePublic,
	}

	_, err := g.Start()
	if err != nil {
		t.Error(err)
	}

	for _, sch := range g.wsdl.Types.Schemas {
		for _, st := range sch.SimpleType {
			if st.List == nil {
				t.Errorf("Nil list for SimpleType")
			}
			if goSimpleType(st) != `[]int32` {
				t.Errorf("SimpleType list was not rendered correctly")
			}
		}
	}

	_, err = g.genTypes()
	if err != nil {
		t.Error(err)
	}
}
