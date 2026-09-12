package snowflake

import (
	"testing"
)

func TestNewNode(t *testing.T) {
	type args struct {
		workerID int64
	}
	tests := []struct {
		name    string
		args    args
		want    *Node
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				workerID: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := NewNode(tt.args.workerID)
			if err != nil {
				t.Errorf("NewNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i := 0; i < 10000; i++ {
				id := node.GetNextID()
				t.Log(id)
			}
		})
	}
}
