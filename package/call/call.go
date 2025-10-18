package call

import "github.com/bsthun/gut"

type Caller interface {
	Call(request *Request, output any) (*Response, *gut.ErrorInstance)
}
