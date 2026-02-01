package config

import "github.com/Sn0wo2/CatSync/config/reader"

func GetDefaultConfig() *Config {
	return &Config{
		Log: Log{
			Level:      reader.Str("debug"),
			Dir:        reader.Str("./logs"),
			FileFormat: reader.Str("2006-01-02.log"),
		},
		Server: Server{
			Address: reader.Str(":3000"),
		},
		Modifiers: []GlobalModifier{
			{
				ActionModifierResponseHeader: &ActionModifierResponseHeader{
					Header: map[string][]string{
						"Server":            {"CatSync"},
						"X-CatSync-Version": {"${VERSION}"},
					},
				},
			},
		},
		Actions: []Action{
			{
				Route: reader.Str("^/$"),
				Type:  ActionString,
				GlobalModifier: GlobalModifier{
					ActionModifierStatus: &ActionModifierStatus{Status: 200},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}},
					},
				},
				ActionString: &ActionStringData{
					GlobalModifier: GlobalModifier{
						ActionModifierVersion: &ActionModifierVersion{Placeholder: reader.Str("${VERSION}")},
					},
					Content: reader.Str("Hello, CatSync!\n\nThis is the default v2 demo config.\n\nTry:\n- /public/hello.txt\n- /demo/headers\n- /demo/status/204\n- /secure (need X-Token: dev)\n- /secure/jump (missing token will jump to a jump-only action)\n\nHint: check response header X-CatSync-Version\n"),
				},
			},
			{
				Route: reader.Str("^/public/hello\\.txt$"),
				Type:  ActionFile,
				GlobalModifier: GlobalModifier{
					ActionModifierStatus: &ActionModifierStatus{Status: 200},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Cache-Control": {"public, max-age=60"}},
					},
				},
				ActionFile: &ActionFileData{
					Path: reader.Str("./data/hello.txt"),
				},
			},
			{
				Route: reader.Str("^/demo/headers$"),
				Type:  ActionString,
				GlobalModifier: GlobalModifier{
					ActionModifierStatus: &ActionModifierStatus{Status: 200},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{
							"Content-Type": {"text/plain; charset=utf-8"},
							"X-Demo":       {"action"},
						},
					},
				},
				ActionString: &ActionStringData{
					GlobalModifier: GlobalModifier{
						ActionModifierResponseHeader: &ActionModifierResponseHeader{
							Header: map[string][]string{"X-Demo": {"payload"}},
						},
					},
					Content: reader.Str("header demo\n\nCheck response headers:\n- global: Server, X-CatSync-Version\n- action: X-Demo=action\n- payload: X-Demo=payload\n\nNote: responseHeader uses Append, so X-Demo will appear multiple times.\n"),
				},
			},
			{
				Route: reader.Str("^/demo/status/204$"),
				Type:  ActionString,
				GlobalModifier: GlobalModifier{
					ActionModifierStatus: &ActionModifierStatus{Status: 204},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}},
					},
				},
				ActionString: &ActionStringData{Content: reader.Str("(status is set to 204 by modifier)\n")},
			},
			{
				Route: reader.Str("^/secure$"),
				Type:  ActionString,
				GlobalModifier: GlobalModifier{
					ActionModifierAuth: &ActionModifierAuth{
						Header: map[string][]*reader.String{
							"X-Token": {reader.Str("^dev$")},
						},
						Fallback: &ActionModifierAuthFallback{Type: AuthFallbackNext},
					},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}},
					},
				},
				ActionString: &ActionStringData{Content: reader.Str("secret ok (X-Token: dev)\n")},
			},
			{
				// Demonstrate AuthFallbackJump: failed auth jumps to a jump-only action.
				Route: reader.Str("^/secure/jump$"),
				Type:  ActionString,
				GlobalModifier: GlobalModifier{
					ActionModifierAuth: &ActionModifierAuth{
						Header: map[string][]*reader.String{
							"X-Token": {reader.Str("^dev$")},
						},
						Fallback: &ActionModifierAuthFallback{Type: AuthFallbackJump, JumpTo: 7},
					},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}},
					},
				},
				ActionString: &ActionStringData{Content: reader.Str("reachable only if auth passes\n")},
			},
			{
				Route: reader.Str("^/secure/jump/fallback$"),
				Type:  ActionString,
				GlobalModifier: GlobalModifier{
					ActionModifierStatus: &ActionModifierStatus{Status: 401},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}},
					},
				},
				ActionString: &ActionStringData{Content: reader.Str("auth failed -> jumped here (route-based fallback action)\n")},
			},
			{
				// The last action is always used as the notfound handler.
				Route: reader.Str(""),
				Type:  ActionString,
				GlobalModifier: GlobalModifier{
					ActionModifierStatus: &ActionModifierStatus{Status: 404},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}},
					},
				},
				ActionString: &ActionStringData{Content: reader.Str("page not found\n")},
			},
		},
	}
}
