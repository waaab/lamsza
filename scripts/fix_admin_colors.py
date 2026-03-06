import re

# 1. Update global.css with new button variables
with open('src/styles/global.css', 'r') as f:
    global_css = f.read()

light_admin_vars = """
  /* Admin Buttons - Light */
  --btn-danger-bg: #dc3545;
  --btn-danger-hover: #c82333;
  --btn-submit-bg: #0d6efd;
  --btn-submit-hover: #0b5ed7;
  --btn-logout-bg: #fd7e14;
  --btn-logout-hover: #e16503;
  --btn-update-bg: #0dcaf0;
  --btn-update-hover: #31d2f2;
"""

dark_admin_vars = """
    /* Admin Buttons - Dark */
    --btn-danger-bg: #e4606d;
    --btn-danger-hover: #dc3545;
    --btn-submit-bg: #3d8bfd;
    --btn-submit-hover: #0d6efd;
    --btn-logout-bg: #fd9843;
    --btn-logout-hover: #fd7e14;
    --btn-update-bg: #3dd5f3;
    --btn-update-hover: #0dcaf0;
"""

# Insert into light root
global_css = global_css.replace('--szekely-red-alpha: rgba(200, 16, 46, 0.2);\n}', '--szekely-red-alpha: rgba(200, 16, 46, 0.2);\n' + light_admin_vars + '}')

# Insert into dark root
global_css = global_css.replace('--szekely-red-alpha: rgba(255, 82, 82, 0.2);\n  }', '--szekely-red-alpha: rgba(255, 82, 82, 0.2);\n' + dark_admin_vars + '  }')

with open('src/styles/global.css', 'w') as f:
    f.write(global_css)

# 2. Update admin.css
with open('src/styles/admin.css', 'r') as f:
    admin_css = f.read()

replacements = {
    '#dc3545': 'var(--btn-danger-bg)',
    '#c82333': 'var(--btn-danger-hover)',
    '#0d6efd': 'var(--btn-submit-bg)',
    '#0b5ed7': 'var(--btn-submit-hover)',
    '#fd7e14': 'var(--btn-logout-bg)',
    '#e16503': 'var(--btn-logout-hover)',
    '#0dcaf0': 'var(--btn-update-bg)',
    '#31d2f2': 'var(--btn-update-hover)',
    'color: #000;': 'color: var(--text-primary);'
}

for old, new in replacements.items():
    admin_css = admin_css.replace(old, new)

with open('src/styles/admin.css', 'w') as f:
    f.write(admin_css)

# 3. Update src/routes/admin/+page.svelte
with open('src/routes/admin/+page.svelte', 'r') as f:
    admin_svelte = f.read()

admin_svelte = admin_svelte.replace('bg_color: "#ffffff"', 'bg_color: "var(--card-bg)"')
admin_svelte = admin_svelte.replace('bg_color: "#ffebd6"', 'bg_color: "var(--warning-bg)"')
admin_svelte = admin_svelte.replace('placeholder="#ffffff"', 'placeholder="var(--card-bg)"')
admin_svelte = admin_svelte.replace('placeholder="#ffebd6"', 'placeholder="var(--warning-bg)"')
admin_svelte = admin_svelte.replace('border:1px solid #ccc;"', 'border:1px solid var(--border-color);"')

with open('src/routes/admin/+page.svelte', 'w') as f:
    f.write(admin_svelte)

# 4. Update src/routes/(public)/hirek/+page.svelte
with open('src/routes/(public)/hirek/+page.svelte', 'r') as f:
    hirek_svelte = f.read()

hirek_svelte = hirek_svelte.replace('feedObj.bg_color || "#ffebd6"', 'feedObj.bg_color || "var(--warning-bg)"')

with open('src/routes/(public)/hirek/+page.svelte', 'w') as f:
    f.write(hirek_svelte)

print("Done fixing inline colors.")
