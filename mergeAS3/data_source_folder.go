package mergeAS3

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/santhosh-tekuri/jsonschema/v3"
	_ "github.com/santhosh-tekuri/jsonschema/v3/httploader"
)

type Metadata struct {
	Tenant string
}

type Tenant struct {
	Path string
	Name string
	Meta Metadata
}

func dataSourceFolders() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFoldersRead,
		Schema: map[string]*schema.Schema{
			"folder": {
				Type:     schema.TypeString,
				Required: true,
			},
			"validation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"schema_version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "3.21.0",
			},
			"schema": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "https://raw.githubusercontent.com/F5Networks/f5-appsvcs-extension/master/schema/3.21.0/as3-schema.json",
			},
			"declarations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tag": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"as3": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceFoldersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	folder := d.Get("folder").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	var result = []map[string]interface{}{}
	var metadata = []Tenant{}

	var parseDir = func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			d := strings.Split(path, "/")
			if len(d) == 3 {
				metadata = append(metadata, Tenant{Path: path, Name: d[2]})
			}
		}

		return nil
	}

	var parseMeta = func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			if filepath.Base(path) == "metainfo.json" {

				var meta Metadata

				data, err := ioutil.ReadFile(path)
				if err != nil {
					diag.FromErr(err)
					return nil
				}
				err = json.Unmarshal(data, &meta)
				if err != nil {
					diag.FromErr(err)
					return nil
				}

				for i := range metadata {
					if metadata[i].Path == filepath.Dir(path) {
						metadata[i].Name = meta.Tenant
						metadata[i].Meta = meta
					}
				}
			}
		}

		return nil
	}

	err := filepath.Walk(folder, parseDir)
	if err != nil {
		return diag.FromErr(err)
	}

	err = filepath.Walk(folder, parseMeta)
	if err != nil {
		return diag.FromErr(err)
	}

	files, err := os.ReadDir(folder)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, file := range files {
		if file.IsDir() {

			var declaration = map[string]interface{}{
				"class":   "AS3",
				"persist": true,
			}

			var as3 = map[string]interface{}{
				"class":         "ADC",
				"schemaVersion": d.Get("schema_version").(string),
				"id":            "something",
			}

			var parseFile = func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() && filepath.Base(path) != "metainfo.json" {
					data, err := ioutil.ReadFile(path)
					if err != nil {
						diag.FromErr(err)
						return nil
					}

					for i := range metadata {
						if metadata[i].Path == filepath.Dir(path) {

							if _, ok := as3[metadata[i].Name]; !ok {
								as3[metadata[i].Name] = map[string]interface{}{
									"class": "Tenant",
								}
							}

							tmp, _ := json.Marshal(as3[metadata[i].Name])

							var tmp2 = map[string]interface{}{}
							json.Unmarshal(tmp, &tmp2)
							err = json.Unmarshal(data, &tmp2)
							if err != nil {
								fmt.Println("File reading error", err)
								return nil
							}

							as3[metadata[i].Name] = tmp2
						}
					}

				}
				return nil
			}

			err = filepath.Walk(folder+"/"+file.Name(), parseFile)
			if err != nil {
				return diag.FromErr(err)
			}

			declaration["declaration"] = as3

			jsonString, err := json.Marshal(&declaration)
			if err != nil {
				return diag.FromErr(err)
			}

			// https://github.com/xeipuuv/gojsonschema
			if d.Get("validation").(bool) {

				schemaReference := d.Get("schema").(string)

				if d.Get("schema_version").(string) != "3.21.0" {
					schemaReference = "https://raw.githubusercontent.com/F5Networks/f5-appsvcs-extension/master/schema/" + d.Get("schema_version").(string) + "/as3-schema.json"
				}

				schema, err := jsonschema.Compile(schemaReference)
				if err != nil {
					return diag.FromErr(err)
				}

				if err = schema.Validate(strings.NewReader(string(jsonString))); err != nil {
					return diag.FromErr(err)
				}
			}

			result = append(result, map[string]interface{}{"tag": file.Name(), "as3": string(jsonString)})
		}
	}

	if err := d.Set("declarations", result); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
