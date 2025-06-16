import { registerCommand } from "../lib/command";

registerCommand({
    name: "test",
    aliases: ["ts"],
    description: "A test command to demonstrate command structure",
    options: [
        {
            name: "option1",
            description: "A required option.",
            required: true,
            exampleValues: ["example1"],
        }
    ],
    flags: [
        {
            name: "flag1",
            description: "A flag which takes no value.",
            short: "f",
            takesValue: false,
        },
        {
            name: "flag2",
            description: "A flag which takes a value.",
            short: "s",
            takesValue: true,
            exampleValues: ["exampleFlag2"],
        }
    ],
    exampleUsage: [
        "minepack test --flag1 --flag2 exampleFlag2Value exampleOption1",
        "minepack test -f -s exampleFlag2Value exampleOption1",
        "minepack test exampleOption1",
    ],
    execute: async ({ flags, options }) => {
        console.log("Executing command with flags:", flags);
        console.log("Executing command with options:", options);
    }
});