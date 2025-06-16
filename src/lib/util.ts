import chalk from "chalk";
import stripAnsi from "strip-ansi";

export interface StatusMessageOptions {
    color?: (msg: string) => string;
    doneMessage?: string;
    doneColor?: (msg: string) => string;
    clearLine?: boolean;
    showDone?: boolean;
    prefix?: string;
    suffix?: string;
}

export function statusMessage(
    message: string,
    options: StatusMessageOptions = {}
): {
    update: (newMessage: string) => void;
    clear: (fn?: () => void) => void;
    done: (doneMsg?: string) => void;
} {
    const stdout = process.stdout;
    const {
        color = chalk.gray,
        doneMessage = " ✔  Done",
        doneColor = chalk.greenBright,
        clearLine = true,
        prefix = "",
        suffix = "",
    } = options;

    function writeMsg(msg: string) {
        if (clearLine) {
            stdout.cursorTo(0);
            stdout.clearLine(1);
        }
        stdout.write(prefix + color(msg) + suffix);
    }

    writeMsg(message);

    return {
        update(newMessage: string) {
            writeMsg(newMessage);
        },
        clear(fn?: () => void) {
            if (fn) fn();
            if (clearLine) {
                stdout.cursorTo(0);
                stdout.clearLine(1);
            }
        },
        done(doneMsg?: string) {
            if (clearLine) {
                stdout.cursorTo(0);
                stdout.clearLine(1);
            }
            stdout.write(
                prefix +
                    doneColor(doneMsg ?? doneMessage) +
                    suffix +
                    "\n"
            );
        },
    };
}

export function selectFromList(list: string[], question: string): Promise<number> {
    return new Promise((resolve) => {
        const stdin = process.stdin;
        const stdout = process.stdout;

        stdout.write(question + "\n");
        list.forEach((item, index) => {
            stdout.write(`${chalk.gray(`${index + 1}.`)} ${item}\n`);
        });
        stdout.write(chalk.blueBright("Select an option (1-" + list.length + ", or type part of an item): "));

        stdin.resume();
        stdin.once("data", (data) => {
            const input = data.toString().trim();

            // Try numeric input first
            const choice = parseInt(input, 10);
            if (!isNaN(choice) && choice >= 1 && choice <= list.length) {
                stdin.pause();
                stdout.write(chalk.greenBright(`You selected: ${list[choice - 1]}\n`));
                resolve(choice - 1); // Return zero-based index
                return;
            }

            // Try exact match (case-insensitive, strip ANSI)
            const exactMatches = list.filter((item) =>
                stripAnsi(item).toLowerCase() === input.toLowerCase()
            );

            if (exactMatches.length === 1) {
                stdin.pause();
                stdout.write(chalk.greenBright(`You selected: ${exactMatches[0]}\n`));
                resolve(list.indexOf(exactMatches[0]));
                return;
            } else if (exactMatches.length > 1) {
                console.error(chalk.yellowBright.bold(" ⚠  Multiple exact matches found. Please be more specific."));
                stdin.pause();
                resolve(selectFromList(list, question));
                return;
            }

            // Try startsWith match (case-insensitive, strip ANSI)
            const startsWithMatches = list.filter((item) =>
                stripAnsi(item).toLowerCase().startsWith(input.toLowerCase())
            );

            if (startsWithMatches.length === 1) {
                stdin.pause();
                stdout.write(chalk.greenBright(`You selected: ${startsWithMatches[0]}\n`));
                resolve(list.indexOf(startsWithMatches[0]));
                return;
            } else if (startsWithMatches.length > 1) {
                console.error(chalk.yellowBright.bold(" ⚠  Multiple matches found (startsWith). Please be more specific."));
                stdin.pause();
                resolve(selectFromList(list, question));
                return;
            }

            // Try includes match (case-insensitive, strip ANSI)
            const includesMatches = list.filter((item) =>
                stripAnsi(item).toLowerCase().includes(input.toLowerCase())
            );

            if (includesMatches.length === 1) {
                stdin.pause();
                stdout.write(chalk.greenBright(`You selected: ${includesMatches[0]}\n`));
                resolve(list.indexOf(includesMatches[0]));
                return;
            } else if (includesMatches.length > 1) {
                console.error(chalk.yellowBright.bold(" ⚠  Multiple matches found (includes). Please be more specific."));
            } else {
                console.error(chalk.redBright.bold(" ✖  No match found. Please try again."));
            }

            stdin.pause();
            // Retry on invalid or ambiguous input
            resolve(selectFromList(list, question));
        });
    });
}

export function promptUser(question: string): Promise<string> {
    return new Promise((resolve) => {
        const stdin = process.stdin;
        const stdout = process.stdout;

        stdin.resume();
        stdout.write(question + " ");

        stdin.once("data", (data) => {
            const answer = data.toString().trim();
            resolve(answer);
            stdin.pause();
        });
    });
}

export async function multiSelectFromList(list: string[], question: string): Promise<number[]> {
    const stdin = process.stdin;
    const stdout = process.stdout;
    let selected: Set<number> = new Set();
    let done = false;
    while (!done) {
        stdout.write("\n" + question + "\n");
        list.forEach((item, index) => {
            const prefix = selected.has(index) ? chalk.greenBright("[x]") : chalk.gray("[ ]");
            stdout.write(`${prefix} ${chalk.gray(`${index + 1}.`)} ${item}\n`);
        });
        stdout.write(chalk.blueBright("Type numbers to toggle (e.g. 1 3 5), 'a' for all, 'd' for done: "));
        stdin.resume();
        const input: string = await new Promise(resolve => stdin.once("data", data => resolve(data.toString().trim())));
        if (input.toLowerCase() === "d" || input.toLowerCase() === "done") {
            done = true;
            break;
        }
        if (input.toLowerCase() === "a" || input.toLowerCase() === "all") {
            selected = new Set(list.map((_, i) => i));
            continue;
        }
        const nums = input.split(/\s+/).map(s => parseInt(s, 10) - 1).filter(i => i >= 0 && i < list.length);
        nums.forEach(i => {
            if (selected.has(i)) selected.delete(i);
            else selected.add(i);
        });
    }
    stdin.pause();
    return Array.from(selected);
}