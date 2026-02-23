package config

import (
	"github.com/Sn0wo2/CatSync/config/reader"
)

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
				Type:  ActionServer,
				GlobalModifier: GlobalModifier{
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Content-Type": {"text/html; charset=utf-8"}},
					},
				},
				ActionServer: &ActionServerData{
					Directory: reader.Str("server/welcome"),
				},
			},
			{
				Route: reader.Str("^/string$"),
				Type:  ActionString,
				GlobalModifier: GlobalModifier{
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}},
					},
				},
				ActionString: &ActionStringData{Content: reader.Str("ActionString: plain text response\n")},
			},
			{
				Route: reader.Str("^/file$"),
				Type:  ActionFile,
				GlobalModifier: GlobalModifier{
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Cache-Control": {"public, max-age=60"}},
					},
				},
				ActionFile: &ActionFileData{Path: reader.Str("./data/hello.txt")},
			},
			{
				Route: reader.Str("^/status/200$"),
				Type:  ActionString,
				GlobalModifier: GlobalModifier{
					ActionModifierStatus: &ActionModifierStatus{Status: 200},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}},
					},
				},
				ActionString: &ActionStringData{Content: reader.Str("Status: 200 OK\n")},
			},
			{
				Route: reader.Str("^/status/201$"),
				Type:  ActionString,
				GlobalModifier: GlobalModifier{
					ActionModifierStatus: &ActionModifierStatus{Status: 201},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}},
					},
				},
				ActionString: &ActionStringData{Content: reader.Str("Status: 201 Created\n")},
			},
			{
				Route: reader.Str("^/status/204$"),
				Type:  ActionString,
				GlobalModifier: GlobalModifier{
					ActionModifierStatus: &ActionModifierStatus{Status: 204},
				},
				ActionString: &ActionStringData{},
			},
			{
				Route: reader.Str("^/headers$"),
				Type:  ActionString,
				GlobalModifier: GlobalModifier{
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{
							"Content-Type": {"text/plain; charset=utf-8"},
							"X-Global":     {"from-global-modifier"},
						},
					},
				},
				ActionString: &ActionStringData{
					GlobalModifier: GlobalModifier{
						ActionModifierResponseHeader: &ActionModifierResponseHeader{
							Header: map[string][]string{"X-Payload": {"from-payload"}},
						},
					},
					Content: reader.Str("Headers demo - check response headers:\n- X-Global: from-global-modifier\n- X-Payload: from-payload\n"),
				},
			},
			{
				Route: reader.Str("^/auth$"),
				Type:  ActionString,
				GlobalModifier: GlobalModifier{
					ActionModifierAuth: &ActionModifierAuth{
						Header: map[string][]*reader.String{
							"X-Token": {reader.Str("^secret$")},
						},
						Fallback: &ActionModifierAuthFallback{Type: AuthFallbackNext},
					},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}},
					},
				},
				ActionString: &ActionStringData{Content: reader.Str("Auth success! (X-Token: secret)\n")},
			},
			{
				Route: reader.Str("^/auth/jump$"),
				Type:  ActionString,
				GlobalModifier: GlobalModifier{
					ActionModifierAuth: &ActionModifierAuth{
						Header: map[string][]*reader.String{
							"X-Token": {reader.Str("^secret$")},
						},
						Fallback: &ActionModifierAuthFallback{Type: AuthFallbackJump, JumpTo: 10},
					},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}},
					},
				},
				ActionString: &ActionStringData{Content: reader.Str("Auth success with jump!\n")},
			},
			{
				Route: reader.Str("^/auth/fallback$"),
				Type:  ActionString,
				GlobalModifier: GlobalModifier{
					ActionModifierStatus: &ActionModifierStatus{Status: 401},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}},
					},
				},
				ActionString: &ActionStringData{Content: reader.Str("Auth failed - jumped to fallback\n")},
			},
			{
				Route: reader.Str("^/reload$"),
				Type:  ActionReload,
				GlobalModifier: GlobalModifier{
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}},
					},
				},
				ActionReload: &ActionReloadData{},
			},
			{
				Route: reader.Str(""),
				Type:  ActionServer,
				GlobalModifier: GlobalModifier{
					ActionModifierStatus: &ActionModifierStatus{Status: 404},
					ActionModifierResponseHeader: &ActionModifierResponseHeader{
						Header: map[string][]string{"Content-Type": {"text/html; charset=utf-8"}},
					},
				},
				ActionServer: &ActionServerData{
					Directory: reader.Str("server/welcome"),
					NotFoundHTML: reader.Str(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>404 - Not Found | CatSync</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            color: #eee;
        }
        .container {
            text-align: center;
            padding: 2rem;
        }
        h1 {
            font-size: 4rem;
            margin: 0;
            color: #ff6b6b;
        }
        p {
            font-size: 1.2rem;
            margin: 1rem 0 2rem;
            color: #aaa;
        }
        .btn {
            display: inline-flex;
            align-items: center;
            gap: 8px;
            padding: 12px 24px;
            background: #4a90d9;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            font-weight: 500;
            transition: background 0.2s;
        }
        .btn:hover {
            background: #357abd;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>404</h1>
        <p>CatSync: Sync the "cat" config backend server</p>
        <a href="https://github.com/Sn0wo2/CatSync" target="_blank" class="btn">
            <svg height="20" viewBox="0 0 16 16" width="20" fill="currentColor">
                <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
            </svg>
            View on GitHub
        </a>
    </div>
</body>
</html>`),
				},
			},
		},
	}
}
