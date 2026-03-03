import re
import sys

def refactor_css(file_path):
    with open(file_path, 'r') as f:
        content = f.read()

    # 1. Replace the entire :root block with our new light theme variables
    light_root = """:root {
  /* Brand Colors */
  --szekely-brown: #7a3c10;
  --szekely-green: #2f4f4f;
  --szekely-red: #c8102e;
  --szekely-blue: #0059b3;
  --warning-orange: #d96600;
  
  /* Theme tokens - Light */
  --bg-body: #f8f8f8;
  --bg-gradient-start: #f8f8f8;
  --bg-gradient-end: #e8e8e8;
  --text-primary: #111111;
  --text-secondary: #595959;
  --text-muted: #737373;
  --text-faint: #757575;
  --text-faintest: #8c8c8c;
  --card-bg: #ffffff;
  --border-color: #d1d1d1;
  --skeleton-bg: #e0e0e0;
  --service-category-bg: #e8e8e8;
  --white: #ffffff;
  
  /* Layout & Components */
  --tab-hover-bg: #fdf5e6;
  --weather-bg: #e6f0ff;
  --warning-bg: #ffebd6;
  
  /* Specific Note Colors */
  --warning-note-bg: rgba(220, 100, 50, 0.06);
  --warning-note-border: rgba(220, 100, 50, 0.4);
  --info-note-bg: rgba(50, 150, 220, 0.07);
  --info-note-border: rgba(50, 150, 220, 0.45);
  --info-box-brown-bg: rgba(50, 150, 220, 0.05);
  
  /* Shadows */
  --shadow-sm: rgba(0, 0, 0, 0.05);
  --shadow-md: rgba(0, 0, 0, 0.08);
  --shadow-lg: rgba(0, 0, 0, 0.14);
  --shadow-xl: rgba(0, 0, 0, 0.15);
  --shadow-btn: rgba(0, 0, 0, 0.2);
  --shadow-btn-hover: rgba(0, 0, 0, 0.25);
  --shadow-sidebar: rgba(0, 0, 0, 0.06);

  /* Overlays / Transparents */
  --overlay-light: rgba(255, 255, 255, 0.05);
  --szekely-red-alpha: rgba(200, 16, 46, 0.2);
}

[data-theme="dark"],
@media (prefers-color-scheme: dark) {
  :root, html:not([data-theme="light"]) {
    /* Brand Colors - Dark */
    --szekely-brown: #d9824c;
    --szekely-green: #4db6ac;
    --szekely-red: #ff5252;
    --szekely-blue: #66b2ff;
    --warning-orange: #ff9933;

    /* Theme tokens - Dark */
    --bg-body: #111111;
    --bg-gradient-start: #111111;
    --bg-gradient-end: #1a1a1a;
    --text-primary: #f0f0f0;
    --text-secondary: #cccccc;
    --text-muted: #a3a3a3;
    --text-faint: #808080;
    --text-faintest: #666666;
    --card-bg: #1e1e1e;
    --border-color: #333333;
    --skeleton-bg: #2a2a2a;
    --service-category-bg: #222222;
    --white: #ffffff; /* keep white for some specific contrasts if needed, or change to #1e1e1e */
    
    /* Layout & Components */
    --tab-hover-bg: #2a2a2a;
    --weather-bg: #1a2a40;
    --warning-bg: #402a1a;
    
    /* Specific Note Colors */
    --warning-note-bg: rgba(255, 153, 51, 0.1);
    --warning-note-border: rgba(255, 153, 51, 0.5);
    --info-note-bg: rgba(102, 178, 255, 0.1);
    --info-note-border: rgba(102, 178, 255, 0.5);
    --info-box-brown-bg: rgba(102, 178, 255, 0.05);

    /* Shadows */
    --shadow-sm: rgba(0, 0, 0, 0.3);
    --shadow-md: rgba(0, 0, 0, 0.4);
    --shadow-lg: rgba(0, 0, 0, 0.5);
    --shadow-xl: rgba(0, 0, 0, 0.6);
    --shadow-btn: rgba(0, 0, 0, 0.7);
    --shadow-btn-hover: rgba(0, 0, 0, 0.8);
    --shadow-sidebar: rgba(0, 0, 0, 0.35);

    /* Overlays / Transparents */
    --overlay-light: rgba(255, 255, 255, 0.05);
    --szekely-red-alpha: rgba(255, 82, 82, 0.2);
  }
}
"""
    # Replace the existing :root
    content = re.sub(r':root\s*\{.*?\n\}\n', light_root, content, flags=re.DOTALL)
    
    # Remove existing dark mode overrides since we combined them
    content = re.sub(r'\[data-theme="dark"\]\s*\{.*?\n\}\n', '', content, flags=re.DOTALL)
    # The [data-theme="dark"] rules for .search-input, .tab:hover, etc. are currently using hex codes.
    # We will replace those hex codes with variables! But wait, if they use variables, they don't need to be inside [data-theme="dark"] pseudo classes, they just inherit from root automatically!
    # So we can just DELETE the block between 866 and 944 entirely.
    
    start_del = content.find('/* ===== Dark Mode Overrides ===== */')
    end_del = content.find('/* ===== /hirek news page layout ===== */')
    if start_del != -1 and end_del != -1:
        content = content[:start_del] + content[end_del:]
        
    # Now replace standard hex & hardcoded values across the entire file

    replacements = [
        # Hex replaces
        (r'#fdf5e6', 'var(--tab-hover-bg)'),
        (r'#007bff', 'var(--szekely-blue)'),
        (r'#e6f0ff', 'var(--weather-bg)'),
        (r'#fd7e14', 'var(--warning-orange)'),
        (r'#ffebd6', 'var(--warning-bg)'),
        (r'#333', 'var(--text-primary)'),
        
        # Shadows replaces
        (r'rgba\(0,\s*0,\s*0,\s*0\.05\)', 'var(--shadow-sm)'),
        (r'rgba\(0,\s*0,\s*0,\s*0\.06\)', 'var(--shadow-sidebar)'),
        (r'rgba\(0,\s*0,\s*0,\s*0\.08\)', 'var(--shadow-md)'),
        (r'rgba\(0,\s*0,\s*0,\s*0\.14\)', 'var(--shadow-lg)'),
        (r'rgba\(0,\s*0,\s*0,\s*0\.15\)', 'var(--shadow-xl)'),
        (r'rgba\(0,\s*0,\s*0,\s*0\.25\)', 'var(--shadow-btn-hover)'),
        (r'rgba\(0,\s*0,\s*0,\s*0\.2\)', 'var(--shadow-btn)'),
        
        # Specific rgba background replacing
        (r'rgba\(255,\s*255,\s*255,\s*0\.05\)', 'var(--overlay-light)'),
        (r'rgba\(220,\s*100,\s*50,\s*0\.06\)', 'var(--warning-note-bg)'),
        (r'rgba\(220,\s*100,\s*50,\s*0\.4\)', 'var(--warning-note-border)'),
        (r'rgba\(50,\s*150,\s*220,\s*0\.07\)', 'var(--info-note-bg)'),
        (r'rgba\(50,\s*150,\s*220,\s*0\.45\)', 'var(--info-note-border)'),
        (r'rgba\(50,\s*150,\s*220,\s*0\.05\)', 'var(--info-box-brown-bg)'),
        (r'rgba\(200,\s*16,\s*46,\s*0\.2\)', 'var(--szekely-red-alpha)'),
        
        # In case we missed #333333 or similar
        (r'#333333', 'var(--text-primary)'),
    ]
    
    for old, new in replacements:
        content = re.sub(old, new, content, flags=re.IGNORECASE)

    # Some variables like rgba in linear-gradient might still be there:
    content = re.sub(r'rgba\(248,\s*249,\s*250,\s*1\)', 'var(--bg-gradient-start)', content)
    content = re.sub(r'rgba\(230,\s*225,\s*220,\s*1\)', 'var(--bg-gradient-end)', content)

    # Some additional logic for [data-theme="dark"] ... .service-tag
    # The original file had:
    # [data-theme="dark"] .service-tag { background: rgba(255, 255, 255, 0.05); }
    # Since we removed [data-theme="dark"] .service-tag, let's just make the .service-tag background use a variable that switches in dark mode.
    # Currently .service-tag is: background: var(--service-category-bg);
    # And in dark mode, we just set --service-category-bg: var(--overlay-light); 
    # Or in dark mode, --service-category-bg: #222222; which works too. So we can just delete lines 405-407 `[data-theme="dark"] .service-tag { ... }`
    content = re.sub(r'\[data-theme="dark"\]\s*\.service-tag\s*\{[^}]+\}', '', content)

    with open(file_path, 'w') as f:
        f.write(content)

if __name__ == "__main__":
    refactor_css("src/styles/global.css")
