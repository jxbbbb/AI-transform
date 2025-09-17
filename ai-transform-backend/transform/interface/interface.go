package _interface

import "context"

type ConsumerTask interface {
	Start(ctx context.Context)
}
