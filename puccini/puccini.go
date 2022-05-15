package puccini

import (
	"fmt"
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/format"
	"github.com/tliron/kutil/logging"
	problemspkg "github.com/tliron/kutil/problems"
	"github.com/tliron/kutil/terminal"
	urlpkg "github.com/tliron/kutil/url"
	cloutpkg "github.com/tliron/puccini/clout"
	"github.com/tliron/puccini/clout/js"
	"github.com/tliron/puccini/tosca/normal"
	"github.com/tliron/puccini/tosca/parser"
	"github.com/tliron/yamlkeys"
	"strings"
)

var log = logging.GetLogger("puccini-server")
var parserContext = parser.NewContext()

func Compile(serviceTemplateUrl, scriptletName, scriptletUrl string, inputs string) (string, error) {
	var url_ urlpkg.URL
	var serviceTemplate *normal.ServiceTemplate
	var clout *cloutpkg.Clout
	var problems *problemspkg.Problems
	var err error

	urlContext := urlpkg.NewContext()
	defer func(urlContext *urlpkg.Context) {
		err := urlContext.Release()
		if err != nil {
			log.Errorf("Failed to close url context: %v", err)
		}
	}(urlContext)

	if url_, err = urlpkg.NewValidURL(serviceTemplateUrl, nil, urlContext); err != nil {
		err = fmt.Errorf("failed to build valid service template url: %v", err)
		log.Error(err.Error())
		return "", err
	}

	inputs_ := make(map[string]ard.Value)
	if inputs != "" {
		if data, err := yamlkeys.DecodeAll(strings.NewReader(inputs)); err == nil {
			for _, data_ := range data {
				if map_, ok := data_.(ard.Map); ok {
					for key, value := range map_ {
						inputs_[yamlkeys.KeyString(key)] = value
					}
				} else {
					err = fmt.Errorf("failed to parse service template inputs: %v", err)
					log.Error(err.Error())
					return "", err
				}
			}
		} else {
			err = fmt.Errorf("failed to parse service template inputs: %v", err)
			log.Error(err.Error())
			return "", err
		}
	}

	if _, serviceTemplate, problems, err = parserContext.Parse(url_, terminal.NewStylist(false), nil, inputs_); err != nil {
		err = fmt.Errorf("failed to parse service template: %v", err)
		log.Error(err.Error())
		return "", err
	}

	if !problems.Empty() {
		err = fmt.Errorf("problems encountered during service template parsing: %v", problems)
		log.Error(err.Error())
		return "", err
	}

	if clout, err = serviceTemplate.Compile(true); err != nil {
		err = fmt.Errorf("failed to compile service template: %v", err)
		log.Error(err.Error())
		return "", err
	}

	js.Resolve(clout, problems, urlContext, true, "", false, false, true)
	if !problems.Empty() {
		err = fmt.Errorf("problems encountered during relationships resolving: %v", problems)
		log.Error(err.Error())
		return "", err
	}

	if scriptletName == "" {
		return format.EncodeYAML(clout, "", false)
	} else {
		scriptlet, err := js.GetScriptlet(scriptletName, clout)

		if err != nil {
			log.Warning("Scriptlet not found")
			// Try loading JavaScript from path or URL
			scriptletUrl, err := urlpkg.NewValidURL(scriptletUrl, nil, urlContext)
			if err != nil {
				err = fmt.Errorf("failed to build valid scriptlet url: %v", err)
				log.Error(err.Error())
				return "", err
			}

			scriptlet, err = urlpkg.ReadString(scriptletUrl)
			if err != nil {
				err = fmt.Errorf("failed to read scriptlet from url: %v", err)
				log.Error(err.Error())
				return "", err
			}

			err = js.SetScriptlet(scriptletName, js.CleanupScriptlet(scriptlet), clout)
			if err != nil {
				err = fmt.Errorf("failed to set scriptlet on service template: %v", err)
				log.Error(err.Error())
				return "", err
			}
		}

		jsContext := js.NewContext(scriptletName, log, map[string]string{}, false, "yaml", false, true, true, "", urlContext)
		var builder strings.Builder
		jsContext.Stdout = &builder
		_, err = jsContext.Require(clout, scriptletName, nil)
		if err != nil {
			err = fmt.Errorf("failed to run scriptlet on service template: %v", err)
			log.Error(err.Error())
			return "", err
		}

		return builder.String(), nil
	}
}
