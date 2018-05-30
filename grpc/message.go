// Copyright 2018 AMIS Technologies
// This file is part of the sol2proto
//
// The sol2proto is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The sol2proto is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the sol2proto. If not, see <http://www.gnu.org/licenses/>.

package grpc

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"sort"
)

func GenerateMessageProtoFile(name, pkgName string, sources []string, messages []Message, version string) MessageProtoFile {
	var filteredMsgs []Message
	var processedSources []string
	encountered := make(map[string]bool)

	for _, m := range messages {
		if _, ok := encountered[m.Name]; !ok {
			encountered[m.Name] = true
			filteredMsgs = append(filteredMsgs, m)
		}
	}

	for _, s := range sources {
		processedSources = append(processedSources, filepath.Base(s))
	}

	sort.Sort(Messages(filteredMsgs))
	sort.Sort(Sources(sources))

	return MessageProtoFile{
		GeneratorVersion: version,
		Package:          pkgName,
		Name:             name,
		Messages:         filteredMsgs,
		Sources:          processedSources,
	}
}

type MessageProtoFile struct {
	GeneratorVersion string
	Package          string
	Name             string
	Messages         Messages
	Sources          Sources
}

func (p MessageProtoFile) Render(writer io.WriteCloser) error {
	template, err := template.New("proto").Parse(MessagesTemplate)
	if err != nil {
		fmt.Printf("Failed to parse template: %v\n", err)
		return err
	}

	return template.Execute(writer, p)
}

var MessagesTemplate string = `// Automatically generated by sol2proto {{ .GeneratorVersion }}. DO NOT EDIT!
// sources: {{ range .Sources }}
//     {{ . }}
{{- end }}
syntax = "proto3";

package {{ .Package }};

import public "github.com/getamis/sol2proto/pb/messages.proto";

{{ range .Messages }}
{{ . }}
{{ end }}
`
