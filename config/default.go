package config

func GetDefaultConfig() *Config {
	return &Config{
		Log: Log{
			Level:      "debug",
			Dir:        "./logs",
			FileFormat: "2006-01-02.log",
		},
		Server: Server{
			Address: ":3000",
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
				Route: "^/$",
				Type:  ActionString,
				GlobalModifier: GlobalModifier{
					ActionModifierStatus: &ActionModifierStatus{Status: 200},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}},
					},
				},
				ActionString: &ActionStringData{
					GlobalModifier: GlobalModifier{
						ActionModifierVersion: &ActionModifierVersion{Placeholder: "${VERSION}"},
					},
					Content: "Hello, CatSync!\n\nThis is the default v2 demo config.\n\nTry:\n- /public/hello.txt\n- /demo/headers\n- /demo/status/204\n- /secure (need X-Token: dev)\n- /secure/jump (missing token will jump to a jump-only action)\n\nHint: check response header X-CatSync-Version\n",
				},
			},
			{
				Route: "^/public/hello\\.txt$",
				Type:  ActionFile,
				GlobalModifier: GlobalModifier{
					ActionModifierStatus: &ActionModifierStatus{Status: 200},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Cache-Control": {"public, max-age=60"}},
					},
				},
				ActionFile: &ActionFileData{
					Path: "./data/hello.txt",
				},
			},
			{
				Route: "^/demo/headers$",
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
					Content: "header demo\n\nCheck response headers:\n- global: Server, X-CatSync-Version\n- action: X-Demo=action\n- payload: X-Demo=payload\n\nNote: responseHeader uses Append, so X-Demo will appear multiple times.\n",
				},
			},
			{
				Route: "^/demo/status/204$",
				Type:  ActionString,
				GlobalModifier: GlobalModifier{
					ActionModifierStatus: &ActionModifierStatus{Status: 204},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}},
					},
				},
				ActionString: &ActionStringData{Content: "(status is set to 204 by modifier)\n"},
			},
			{
				Route: "^/secure$",
				Type:  ActionString,
				GlobalModifier: GlobalModifier{
					ActionModifierAuth: &ActionModifierAuth{
						Header: map[string][]string{
							"X-Token": {"^dev$"},
						},
						Fallback: &ActionModifierAuthFallback{Type: AuthFallbackNext},
					},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}},
					},
				},
				ActionString: &ActionStringData{Content: "secret ok (X-Token: dev)\n"},
			},
			{
				// Demonstrate AuthFallbackJump: failed auth jumps to a jump-only action.
				Route: "^/secure/jump$",
				Type:  ActionString,
				GlobalModifier: GlobalModifier{
					ActionModifierAuth: &ActionModifierAuth{
						Header: map[string][]string{
							"X-Token": {"^dev$"},
						},
						Fallback: &ActionModifierAuthFallback{Type: AuthFallbackJump, JumpTo: 7},
					},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}},
					},
				},
				ActionString: &ActionStringData{Content: "reachable only if auth passes\n"},
			},
			{
				Route: "^/secure/jump/fallback$",
				Type:  ActionString,
				GlobalModifier: GlobalModifier{
					ActionModifierStatus: &ActionModifierStatus{Status: 401},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}},
					},
				},
				ActionString: &ActionStringData{Content: "auth failed -> jumped here (route-based fallback action)\n"},
			},
			{
				// The last action is always used as the notfound handler.
				Route: "",
				Type:  ActionString,
				GlobalModifier: GlobalModifier{
					ActionModifierStatus: &ActionModifierStatus{Status: 404},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}},
					},
				},
				ActionString: &ActionStringData{Content: "page not found\n"},
			},
		},
	}
}
