package main

//import "fmt"

// TODO should be protobuf of the whole table field.type
//func (f *fieldType) jsonV1(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
//	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
//		return res
//	}
//
//	if by, err := json.Marshal(fieldTypes); err == nil {
//		b.Write(by)
//	} else {
//		return weft.InternalServerError(err)
//	}
//
//	h.Set("Content-Type", "application/json;version=1")
//
//	return &weft.StatusOK
//}
