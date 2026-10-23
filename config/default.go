package config

import "github.com/Sn0wo2/CatSync/config/reader"

const (
	contentTypeHeader    = "Content-Type"
	htmlContentType      = "text/html; charset=utf-8"
	plainTextContentType = "text/plain; charset=utf-8"
)

const DefaultNotFoundHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>404 | CatSync</title>
    <style>
        *{margin:0;padding:0;box-sizing:border-box}
        body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',system-ui,sans-serif;display:flex;justify-content:center;align-items:center;min-height:100vh;background:#0a0a0a;color:#e0e0e0}
        .c{text-align:center;padding:2rem}
        .code{font-size:8rem;font-weight:200;letter-spacing:-.04em;line-height:1;color:#fff}
        .line{width:48px;height:1px;background:#333;margin:1.5rem auto}
        .msg{font-size:.875rem;color:#666;letter-spacing:.08em;text-transform:uppercase}
        .link{display:inline-block;margin-top:2.5rem;font-size:.75rem;color:#555;text-decoration:none;letter-spacing:.06em;border-bottom:1px solid #333;padding-bottom:2px;transition:color .2s,border-color .2s}
        .link:hover{color:#fff;border-color:#fff}
    </style>
</head>
<body>
    <div class="c">
        <div class="code">404</div>
        <div class="line"></div>
        <p class="msg">Not Found</p>
        <a href="https://github.com/Sn0wo2/CatSync" class="link" target="_blank" rel="noopener">CatSync</a>
    </div>
</body>
</html>`

func GetDefaultConfig() *Config {
	return &Config{
		Log: Log{
			Level:      &reader.String{Content: "debug"},
			Dir:        &reader.String{Content: "./logs"},
			FileFormat: &reader.String{Content: "2006-01-02.log"},
		},
		Server: Server{
			Address: &reader.String{Content: ":3000"},
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
				Route: &reader.String{Content: "^/$"},
				Type:  &reader.String{Content: "server"},
				GlobalModifier: GlobalModifier{
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{contentTypeHeader: {htmlContentType}},
					},
				},
				ActionServer: &ActionServerData{
					Directory: &reader.String{Content: "server/welcome"},
				},
			},
			{
				Route: &reader.String{Content: "^/string$"},
				Type:  &reader.String{Content: string(ActionString)},
				GlobalModifier: GlobalModifier{
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{contentTypeHeader: {plainTextContentType}},
					},
				},
				ActionString: &ActionStringData{Content: &reader.String{Content: "ActionString: plain text response\n"}},
			},
			{
				Route: &reader.String{Content: "^/file$"},
				Type:  &reader.String{Content: "file"},
				GlobalModifier: GlobalModifier{
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Cache-Control": {"public, max-age=60"}},
					},
				},
				ActionFile: &ActionFileData{Path: &reader.String{Content: "./data/hello.txt"}},
			},
			{
				Route: &reader.String{Content: "^/status/200$"},
				Type:  &reader.String{Content: string(ActionString)},
				GlobalModifier: GlobalModifier{
					ActionModifierStatus: &ActionModifierStatus{Status: 200},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{contentTypeHeader: {plainTextContentType}},
					},
				},
				ActionString: &ActionStringData{Content: &reader.String{Content: "Status: 200 OK\n"}},
			},
			{
				Route: &reader.String{Content: "^/status/201$"},
				Type:  &reader.String{Content: string(ActionString)},
				GlobalModifier: GlobalModifier{
					ActionModifierStatus: &ActionModifierStatus{Status: 201},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{contentTypeHeader: {plainTextContentType}},
					},
				},
				ActionString: &ActionStringData{Content: &reader.String{Content: "Status: 201 Created\n"}},
			},
			{
				Route: &reader.String{Content: "^/status/204$"},
				Type:  &reader.String{Content: string(ActionString)},
				GlobalModifier: GlobalModifier{
					ActionModifierStatus: &ActionModifierStatus{Status: 204},
				},
				ActionString: &ActionStringData{},
			},
			{
				Route: &reader.String{Content: "^/headers$"},
				Type:  &reader.String{Content: string(ActionString)},
				GlobalModifier: GlobalModifier{
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{
							contentTypeHeader: {plainTextContentType},
							"X-Global":        {"from-global-modifier"},
						},
					},
				},
				ActionString: &ActionStringData{
					GlobalModifier: GlobalModifier{
						ActionModifierResponseHeader: &ActionModifierResponseHeader{
							Header: map[string][]string{"X-Payload": {"from-payload"}},
						},
					},
					Content: &reader.String{Content: "Headers demo - check response headers:\n- X-Global: from-global-modifier\n- X-Payload: from-payload\n"},
				},
			},
			{
				Route: &reader.String{Content: "^/auth$"},
				Type:  &reader.String{Content: string(ActionString)},
				GlobalModifier: GlobalModifier{
					ActionModifierAuth: &ActionModifierAuth{
						Header: map[string][]*reader.String{
							"X-Token": {{Content: "^secret$"}},
						},
						Fallback: &ActionModifierAuthFallback{Type: AuthFallbackNext},
					},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{contentTypeHeader: {plainTextContentType}},
					},
				},
				ActionString: &ActionStringData{Content: &reader.String{Content: "Auth success! (X-Token: secret)\n"}},
			},
			{
				Route: &reader.String{Content: "^/auth/jump$"},
				Type:  &reader.String{Content: string(ActionString)},
				GlobalModifier: GlobalModifier{
					ActionModifierAuth: &ActionModifierAuth{
						Header: map[string][]*reader.String{
							"X-Token": {{Content: "^secret$"}},
						},
						Fallback: &ActionModifierAuthFallback{Type: AuthFallbackJump, JumpTo: 10},
					},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{contentTypeHeader: {plainTextContentType}},
					},
				},
				ActionString: &ActionStringData{Content: &reader.String{Content: "Auth success with jump!\n"}},
			},
			{
				Route: &reader.String{Content: "^/auth/fallback$"},
				Type:  &reader.String{Content: string(ActionString)},
				GlobalModifier: GlobalModifier{
					ActionModifierStatus: &ActionModifierStatus{Status: 401},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{contentTypeHeader: {plainTextContentType}},
					},
				},
				ActionString: &ActionStringData{Content: &reader.String{Content: "Auth failed - jumped to fallback\n"}},
			},
			{
				Route: &reader.String{Content: "^/reload$"},
				Type:  &reader.String{Content: "reload"},
				GlobalModifier: GlobalModifier{
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{contentTypeHeader: {plainTextContentType}},
					},
				},
				ActionReload: &ActionReloadData{},
			},
			{
				Route: &reader.String{},
				Type:  &reader.String{Content: "server"},
				GlobalModifier: GlobalModifier{
					ActionModifierStatus: &ActionModifierStatus{Status: 404},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{contentTypeHeader: {htmlContentType}},
					},
				},
				ActionServer: &ActionServerData{
					Directory:    &reader.String{Content: "server/welcome"},
					NotFoundHTML: &reader.String{Content: DefaultNotFoundHTML},
				},
			},
		},
	}
}
