/* eslint-disable no-undef */
// Builds src/views/*.html into dist/ with posthtml-modules + htmlnano.
// Replaces posthtml-cli, whose update-notifier dependency pulls a vulnerable got@9.
const fs = require('node:fs');
const path = require('node:path');
const posthtml = require('posthtml');
const posthtmlModules = require('posthtml-modules');
const htmlnano = require('htmlnano');

const viewsDir = 'src/views';
const outputDir = 'dist';

async function build() {
    fs.mkdirSync(outputDir, { recursive: true });
    const files = fs.readdirSync(viewsDir).filter((file) => file.endsWith('.html'));
    for (const file of files) {
        const from = path.join(viewsDir, file);
        const html = fs.readFileSync(from, 'utf8');
        const result = await posthtml([
            posthtmlModules({ root: './src/views', initial: true }),
            htmlnano({}),
        ]).process(html, { from });
        fs.writeFileSync(path.join(outputDir, file), result.html);
        console.log(`The file ${from} has been saved!`);
    }
}

build().catch((error) => {
    console.error(error);
    process.exit(1);
});
