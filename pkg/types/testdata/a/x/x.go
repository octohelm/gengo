package x

import "context"

func Do[Data any, Op interface{ ResponseData() *Data }](ctx context.Context, req Op) (*Data, error) {
	return nil, nil
}
