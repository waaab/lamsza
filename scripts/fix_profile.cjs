const fs = require('fs');
const file = 'src/routes/(public)/bejegyzes/[slug]/+page.svelte';
let content = fs.readFileSync(file, 'utf8');

const newContent = `            <!-- Remodeled Full-Width Profile -->
            <div style="margin-bottom: 2rem;">
                <div class="badge" style="font-size: 1rem; margin-bottom: 1rem; display: inline-block;">
                    Index: {entry.category}
                </div>
                <h1 style="margin: 0 0 1rem; font-size: 2.5rem; color: var(--text-color);">
                    {entry.name}
                </h1>
                
                {#if entry.url}
                    <div style="margin-bottom: 1.5rem; font-size: 1.1rem;">
                        <span style="color: var(--text-faint); margin-right: 0.5rem;">🔗 Weboldal:</span>
                        <a href={entry.url} target="_blank" rel="nofollow noopener" style="color: var(--primary-color); text-decoration: none; font-weight: 500;">
                            {entry.url}
                        </a>
                    </div>
                {/if}
            </div>

            <!-- Distinct Contact Block -->
            <div style="padding: 1.5rem; background: var(--card-bg); border-radius: 12px; border: 1px solid var(--border-color); margin-bottom: 2rem;">
                <h3 style="margin-top: 0; color: var(--text-color); font-size: 1.2rem; margin-bottom: 1rem;">Kapcsolat</h3>
                <div style="display: grid; gap: 1rem;">
                    {#if entry.location || entry.address}
                        <div style="display: flex; align-items: flex-start; gap: 0.8rem; font-size: 1.1rem;">
                            <span style="font-size: 1.3rem;">📍</span>
                            <div>
                                {#if entry.location}
                                    <strong>{entry.location}</strong>
                                {/if}
                                {#if entry.location && entry.address} - {/if}
                                {#if entry.address}{entry.address}{/if}
                            </div>
                        </div>
                    {/if}
                    {#if entry.phone}
                        <div style="display: flex; align-items: center; gap: 0.8rem; font-size: 1.1rem;">
                            <span style="font-size: 1.3rem;">📞</span>
                            <a href=\`tel:\${entry.phone.replace(/[^0-9+]/g, "")}\` style="color: var(--text-color); text-decoration: none;">{entry.phone}</a>
                        </div>
                    {/if}
                </div>
            </div>

            <!-- Details Block -->
            <div style="padding: 1.5rem; background: var(--bg-body); border-radius: 12px; margin-bottom: 2rem;">
                <h3 style="margin-top: 0; color: var(--text-color); font-size: 1.2rem; margin-bottom: 1rem;">Részletek & Megjegyzések</h3>
                {#if entry.notes}
                    <div class="entry-notes" style="font-size: 1.05rem; line-height: 1.6; margin-bottom: 1.5rem; white-space: pre-wrap; color: var(--text-faint);">
                        {entry.notes}
                    </div>
                {/if}
                
                {#if entry.tags}
                    <div class="entry-tags">
                        {#each entry.tags.split(" ") as t}
                            <span class="entry-tag" style="padding: 0.4rem 0.8rem; font-size: 0.95rem;">{t.startsWith("#") ? t : "#" + t}</span>
                        {/each}
                    </div>
                {/if}
            </div>`;

const startIndex = content.indexOf('<article');
const endIndex = content.indexOf('</article>') + '</article>'.length;

if (startIndex !== -1 && endIndex !== -1) {
    content = content.substring(0, startIndex) + newContent + content.substring(endIndex);
    fs.writeFileSync(file, content);
    console.log("File rewritten");
} else {
    console.error("Could not find <article> marks");
}
