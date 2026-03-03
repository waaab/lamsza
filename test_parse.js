const fs = require('fs');
const { DOMParser } = require('xmldom');

async function test() {
    const res = await fetch('http://localhost:3000/api/proxy?url=https%3A%2F%2Fszekelyhon.ro%2Frss%2Fszekelyhon_hirek.xml');
    const text = await res.text();
    console.log("Length:", text.length);
    const parser = new DOMParser();
    const doc = parser.parseFromString(text, 'text/xml');
    const items = doc.getElementsByTagName("item");
    console.log("Items found:", items.length);
}
test().catch(console.error);
