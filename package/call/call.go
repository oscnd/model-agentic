package call

import "github.com/bsthun/gut"

type Caller interface {
	Call(request *Request, option *Option, output any) (*Response, *gut.ErrorInstance)
}
