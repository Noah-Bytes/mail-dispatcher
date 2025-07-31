package services

import (
	"testing"
)

func TestParseSubject(t *testing.T) {
	service := &MailRoutingService{}

	tests := []struct {
		name        string
		subject     string
		wantKeyword string
		wantTarget  string
		wantErr     bool
	}{
		{
			name:        "正常格式",
			subject:     "报警 - 张三",
			wantKeyword: "报警",
			wantTarget:  "张三",
			wantErr:     false,
		},
		{
			name:        "带空格",
			subject:     " 通知 - 财务部 ",
			wantKeyword: "通知",
			wantTarget:  "财务部",
			wantErr:     false,
		},
		{
			name:        "格式错误",
			subject:     "报警张三",
			wantKeyword: "",
			wantTarget:  "",
			wantErr:     true,
		},
		{
			name:        "空主题",
			subject:     "",
			wantKeyword: "",
			wantTarget:  "",
			wantErr:     true,
		},
		{
			name:        "只有关键字",
			subject:     "报警 - ",
			wantKeyword: "",
			wantTarget:  "",
			wantErr:     true,
		},
		{
			name:        "只有目标",
			subject:     " - 张三",
			wantKeyword: "",
			wantTarget:  "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyword, target, err := service.parseSubject(tt.subject)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseSubject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if keyword != tt.wantKeyword {
				t.Errorf("parseSubject() keyword = %v, want %v", keyword, tt.wantKeyword)
			}

			if target != tt.wantTarget {
				t.Errorf("parseSubject() target = %v, want %v", target, tt.wantTarget)
			}
		})
	}
}
