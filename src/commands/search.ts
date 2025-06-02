import chalk from "chalk";
import { Command, registerCommand } from "../lib/command";

const MODRINTH_API = "https://api.modrinth.com/v2";

async function searchModrinth(query: string) {
    const url = `${MODRINTH_API}/search?query=${encodeURIComponent(query)}&limit=5`;
    const res = await fetch(url);
    if (!res.ok) throw new Error("Failed to search Modrinth");
    const data = await res.json();
    return data.hits || [];
}

const searchCommand = new Command({
    name: "search",
    description: "Search Modrinth for mods and return their Modrinth page URLs.",
    arguments: [
        { name: "query", aliases: [], description: "Search term for Modrinth", required: true }
    ],
    flags: [],
    examples: [
        { description: "Search for a mod on Modrinth", usage: "minepack search sodium" }
    ],
    async execute(args) {
        const query = args.query as string;
        console.log(chalk.gray(`[info] Searching Modrinth for: ${query}`));
        const results = await searchModrinth(query);
        if (!results.length) {
            console.log(chalk.red("No mods found for that search term."));
            return;
        }
        console.log(chalk.bold("Top 5 Modrinth results:"));
        results.forEach((r: any, i: number) => {
            const url = `https://modrinth.com/mod/${r.slug}`;
            console.log(`  [${i + 1}] ${chalk.green(r.title)} (${r.project_id}) - ${r.description}`);
            console.log(`      ${chalk.cyan(url)}`);
        });
    }
});

registerCommand(searchCommand);

export { searchCommand };
