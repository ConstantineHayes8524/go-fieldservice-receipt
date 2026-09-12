package main

import "testing"

func TestShouldSendReceipt(t *testing.T) {
	tests := []struct {
		name, status, email string
		want                bool
	}{{"completed customer", "completed", "a@example.com", true}, {"dispatched", "dispatched", "a@example.com", false}, {"completed no address", "completed", "", false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldSendReceipt(WorkOrder{Status: tt.status, CustomerEmail: tt.email}); got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}
