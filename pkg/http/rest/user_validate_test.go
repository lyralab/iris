package rest

import "testing"

func TestTelegramIDAcceptsUserAndChannel(t *testing.T) {
	cases := []struct {
		id string
		ok bool
	}{
		{id: "", ok: true},
		{id: "123456789", ok: true},
		{id: "-1001234567890", ok: true},
		{id: "@channel", ok: false},
		{id: "12-3", ok: false},
	}
	for _, tc := range cases {
		body := UserUpdateBody{
			UserName:     "admin",
			TelegramID:   tc.id,
			MattermostID: "mm-user-or-channel",
			BaleID:       tc.id,
		}
		err := validate.Struct(body)
		if tc.ok && err != nil {
			t.Errorf("telegram_id %q: %v", tc.id, err)
		}
		if !tc.ok && err == nil {
			t.Errorf("telegram_id %q: expected validation error", tc.id)
		}
	}
}
