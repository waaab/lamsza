import fs from 'fs';
import path from 'path';
import crypto from 'crypto';

const distDir = path.resolve('dist');

// Recursively find all HTML files
function getHtmlFiles(dir, fileList = []) {
    const files = fs.readdirSync(dir);
    for (const file of files) {
        const filePath = path.join(dir, file);
        if (fs.statSync(filePath).isDirectory()) {
            getHtmlFiles(filePath, fileList);
        } else if (filePath.endsWith('.html')) {
            fileList.push(filePath);
        }
    }
    return fileList;
}

const htmlFiles = getHtmlFiles(distDir);

htmlFiles.forEach((file) => {
    let content = fs.readFileSync(file, 'utf-8');

    // Find <script> tags with content (not just src attributes)
    // This matches the typical SvelteKit inline script block
    const scriptRegex = /<script(?!\s+src=)[^>]*>([\s\S]*?)<\/script>/gi;
    let match;
    let hasChanges = false;

    while ((match = scriptRegex.exec(content)) !== null) {
        const fullTag = match[0];
        let scriptContent = match[1].trim();

        if (scriptContent.length > 0) {
            // Fix 1: Module scripts run in strict mode, so undeclared variable assignments throw a ReferenceError.
            // Convert "__sveltekit_... =" to "window.__sveltekit_... ="
            scriptContent = scriptContent.replace(/__sveltekit_[a-zA-Z0-9_]+\s*=/g, (match) => `window.${match}`);

            // Fix 2: document.currentScript is null in type="module" scripts.
            // Replace it with document.body.firstElementChild (which is what SvelteKit targets)
            scriptContent = scriptContent.replace(
                /document\.currentScript\.parentElement/g,
                'document.body.firstElementChild'
            );

            // Determine the path to save the script based on where the HTML file is
            const scriptDirPath = path.join(path.dirname(file), 'app', 'extracted-scripts');

            // Fix 3: Adjust relative import paths since the script is moved into a subdirectory
            const htmlDir = path.dirname(file);
            const relativePathToHtmlDir = path.relative(scriptDirPath, htmlDir).replace(/\\/g, '/');
            scriptContent = scriptContent.replace(/import\("(\.[^"]+)"\)/g, (match, p1) => {
                const newPath = path.posix.join(relativePathToHtmlDir, p1);
                return `import("${newPath.startsWith('.') ? newPath : './' + newPath}")`;
            });

            // Generate a hash for the filename to ensure uniqueness and cache-busting
            const hash = crypto.createHash('sha256').update(scriptContent).digest('hex').substring(0, 8);
            const scriptFilename = `inline-script-${hash}.js`;

            if (!fs.existsSync(scriptDirPath)) {
                fs.mkdirSync(scriptDirPath, { recursive: true });
            }

            const scriptPath = path.join(scriptDirPath, scriptFilename);
            fs.writeFileSync(scriptPath, scriptContent);

            // Calculate relative path from the HTML file to the new script file
            const relativePathToScript = path.relative(path.dirname(file), scriptPath).replace(/\\/g, '/');

            // Replace the inline script with a script tag pointing to the new file
            // Preserve type="module" if it was present
            const typeMatch = fullTag.match(/type="([^"]+)"/);
            const typeAttr = typeMatch ? ` type="${typeMatch[1]}"` : ' type="module"'; // Default to module for SvelteKit

            const newTag = `<script${typeAttr} src="${relativePathToScript}"></script>`;
            content = content.replace(fullTag, newTag);
            hasChanges = true;
        }
    }

    if (hasChanges) {
        fs.writeFileSync(file, content);
        console.log(`Extracted inline scripts from ${path.relative(process.cwd(), file)}`);
    }
});

console.log('Inline script extraction completed.');
