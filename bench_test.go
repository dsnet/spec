package spec

import (
	jsonv1 "encoding/json"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/dsnet/try"
	jsonv2 "github.com/go-json-experiment/json"
	"github.com/google/go-cmp/cmp"
)

var content = func() []byte {
	resp := try.E1(http.Get("https://raw.githubusercontent.com/kubernetes/kubernetes/38e99289d6179672b3fb38be917ed49669f8b719/api/openapi-spec/swagger.json"))
	defer resp.Body.Close()
	return try.E1(io.ReadAll(resp.Body))
}()

func init() {
	// Sanity check that v1 and v2 behavior are identical.
	var v1, v2 Swagger
	try.E(jsonv1.Unmarshal(content, &v1))
	try.E(jsonv2.Unmarshal(content, &v2))
	exportAny := cmp.Exporter(func(reflect.Type) bool { return true })
	if diff := cmp.Diff(v1, v2, exportAny); diff != "" {
		panic(diff)
	}
}

func BenchmarkUnmarshalJSON(b *testing.B) {
	b.Run("V1/Swagger", benchmark[Swagger](jsonv1.Unmarshal))
	b.Run("V1/Interface", benchmark[any](jsonv1.Unmarshal))
	b.Run("V2/Swagger", benchmark[Swagger](jsonv2.Unmarshal))
	b.Run("V2/Interface", benchmark[any](jsonv2.Unmarshal))
}

func benchmark[T any](unmarshal func([]byte, any) error) func(b *testing.B) {
	return func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			try.E(unmarshal(content, new(T)))
		}
	}
}
