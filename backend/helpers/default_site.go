package helpers

import "fmt"

func defaultSiteHTML(domain string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to MiStatic</title>
    <style>
        body {
            font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background-color: #f6f8fa;
            color: #24292f;
        }
        .container {
            text-align: center;
            background: white;
            padding: 3rem;
            border-radius: 12px;
            box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
            max-width: 500px;
        }
        h1 { color: #2da44e; margin-bottom: 0.5rem; }
        p { color: #57606a; margin-bottom: 2rem; line-height: 1.5; }
        .domain { font-family: monospace; background: #f6f8fa; padding: 0.2rem 0.4rem; border-radius: 4px; border: 1px solid #d0d7de; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Ready to Launch!</h1>
        <p>The site for <span class="domain">%s</span> has been successfully created, but no files have been deployed yet.</p>
        <p style="font-size: 0.9em;">Head over to your MiStatic dashboard and upload your first .zip file to get started.</p>
    </div>
</body>
</html>`, domain)
}
