with open('src/styles/global.css', 'r') as f:
    text = f.read()

# We need to find `[data-theme="dark"],\n@media (prefers-color-scheme: dark) {\n\n  :root,\n  html:not([data-theme="light"]) {`
# And replace it with a valid structure. We'll extract the core variables.

import re

# Extract everything between `html:not([data-theme="light"]) {` and the corresponding `  }\n}`
pattern = r'\[data-theme="dark"\],\s*@media \(prefers-color-scheme: dark\)\s*\{\s*:root,\s*html:not\(\[data-theme="light"\]\)\s*\{([\s\S]*?)\s*\}\s*\}'

match = re.search(pattern, text)
if match:
    variables = match.group(1)
    
    valid_css = f"""
[data-theme="dark"] {{
{variables}
}}

@media (prefers-color-scheme: dark) {{
  :root:not([data-theme="light"]) {{
{variables}
  }}
}}
"""
    new_text = text[:match.start()] + valid_css.strip() + text[match.end():]
    with open('src/styles/global.css', 'w') as f:
        f.write(new_text)
    print("Syntax fixed successfully!")
else:
    print("Pattern not found!")

