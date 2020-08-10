package types

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultiError_Add(t *testing.T) {
	type fields struct {
		errs   []error
		errCnt int
	}
	type args struct {
		err error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "case 0",
			fields: fields{
				errs:   []error{},
				errCnt: 0,
			},
			args: args{err: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &MultiError{
				errs:   tt.fields.errs,
				errCnt: tt.fields.errCnt,
			}
			_ = c
		})
	}
}

func TestMultiError_Error(t *testing.T) {
	err := errors.New("test error")

	type fields struct {
		errs   []error
		errCnt int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "case 0",
			fields: fields{
				errs:   []error{err},
				errCnt: 1,
			},
			want: err.Error() + ";",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &MultiError{
				errs:   tt.fields.errs,
				errCnt: tt.fields.errCnt,
			}
			if got := c.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultiError_Nil(t *testing.T) {
	type fields struct {
		errs   []error
		errCnt int
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "case 0",
			fields: fields{
				errs:   nil,
				errCnt: 0,
			},
			want: true,
		},
		{
			name: "case 1",
			fields: fields{
				errs:   []error{errors.New("haha")},
				errCnt: 1,
			},
			want: false,
		},
		{
			name: "case 2",
			fields: fields{
				errs:   []error{errors.New("haha")},
				errCnt: 0,
			},
			want: false,
		},
		{
			name: "case 2",
			fields: fields{
				errs:   []error{},
				errCnt: 1,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &MultiError{
				errs:   tt.fields.errs,
				errCnt: tt.fields.errCnt,
			}
			if got := c.Nil(); got != tt.want {
				t.Errorf("Nil() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultiError_Reset(t *testing.T) {
	type fields struct {
		errs   []error
		errCnt int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "case 0",
			fields: fields{
				errs:   []error{errors.New("haha")},
				errCnt: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &MultiError{
				errs:   tt.fields.errs,
				errCnt: tt.fields.errCnt,
			}

			c.Reset()
			assert.Empty(t, c)
		})
	}
}
