package console

import (
	"testing"

	"github.com/danifv27/soup/internal/application/notification"
	"github.com/stretchr/testify/assert"
)

func TestConsoleNotificationService_Notify(t *testing.T) {
	type args struct {
		notification notification.Notification
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Should not return error",
			args: args{
				notification: notification.Notification{
					Message: "Test Message",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			co := NotificationService{}
			err := co.Notify(tt.args.notification)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
