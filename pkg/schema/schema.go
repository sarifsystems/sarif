package schema

import "reflect"

type Thing struct {
	SchemaContext string `json:"@context,omitempty"`
	SchemaType    string `json:"@type,omitempty"`
	SchemaId      string `json:"@id,omitempty"`
	SchemaLabel   string `json:"rdfs:label,omitempty"`
}

var (
	pkgPath string
)

func Fill(thing interface{}) interface{} {
	v := reflect.ValueOf(thing)
	if v.Kind() != reflect.Ptr {
		panic("schema.Fill: expects struct pointer")
	}

	v = reflect.Indirect(v)
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return thing
	}

	vthing := v.FieldByName("Thing")
	if !vthing.IsValid() {
		return thing
	}

	if vthing.Kind() == reflect.Ptr {
		vthing = reflect.Indirect(vthing)
		if !vthing.IsValid() {
			vthing = reflect.ValueOf(&Thing{})
			v.FieldByName("Thing").Set(vthing)
			vthing = reflect.Indirect(vthing)
		}
	}
	vtype := vthing.FieldByName("SchemaType")
	if vtype.String() != "" {
		return thing
	}

	if path := t.PkgPath(); path != pkgPath {
		fthing, _ := t.FieldByName("Thing")
		tag := fthing.Tag.Get("schema")
		if tag != "" {
			vtype.SetString(tag)
		} else {
			vtype.SetString("http://" + path + "/" + t.Name())
		}
	} else {
		vtype.SetString(t.Name())
	}

	return thing
}

func init() {
	pkgPath = reflect.TypeOf(Thing{}).PkgPath()
}
