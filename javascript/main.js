import { spawnSync } from "node:child_process";
import { createRequire } from "node:module";
import { ArgumentRewriter } from "./args.js";

export function main() {
    callCmd(process.argv.slice(2));
}

export function callCmd(args) {
    const binaryPath = getBinaryPath();
    if (!binaryPath) {
        console.error("Unsupported platform or architecture");
        process.exit(1);
    }
    const rewriter = new ArgumentRewriter();
    const adjusted_args = rewriter.rewrite(args);
    const result = spawnSync(binaryPath, adjusted_args, {
        shell: false,
        stdio: "inherit",
    });
    if (result.error) {
        throw result.error;
    }
    process.exitCode = result.status || 0;
}

function getBinaryPath() {
    const DODO_BINARY_PATH = process.env.DODO_BINARY_PATH;
    if (DODO_BINARY_PATH) {
        return DODO_BINARY_PATH;
    }
    const availablePlatforms = ["linux", "darwin"];
    const availableArchs = ["x86-64", "arm64"];
    const { platform, arch } = process;
    if (!availablePlatforms.includes(platform) || !availableArchs.includes(arch)) {
        return;
    }
    const binaryName = `@dodo/cli-${platform}-${arch}/dodo-cli`;
    const require = createRequire(import.meta.url);
    return require.resolve(binaryName);
}