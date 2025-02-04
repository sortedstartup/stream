const { execSync } = require ('child_process');
const fs = require('fs');
const path = require('path');

console.log("running get-git-hash.cjs")
// Get the Git hash
const gitHash = execSync('git rev-parse --short HEAD').toString().trim();

// Create a timestamp
// read contents of file version.number and store in version_number variable
const version_number = fs.readFileSync(path.join(__dirname, 'version.number'), 'utf8');

// const version_number= "0.0.1"

// Create the version info
const versionInfo = `
// This file is auto-generated. Do not edit manually.
export const VERSION_INFO = {
  gitHash: '${gitHash}',
  version: '${version_number}'
};
`;

// Write to a TypeScript file
fs.writeFileSync(path.join(__dirname, 'versionInfo.ts'), versionInfo);

console.log('Version info written to src/version/versionInfo.ts');