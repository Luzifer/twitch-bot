---
title: Twitch-Bot Documentation
bookToC: false
---

> [!TIP]
> You are tired of all those cloud-bots working only sometimes, messing up at random times (always when you need them most) and limiting you to very little functionality while forcing you to give them every single possible permission available?
>
> Twitch-Bot is a **fully open-source** bot you can host yourself giving you **full control** over its functions and being **extensible** using custom scripts and commands any developer can build for you. Additionally you are in control which **permissions** to give to the bot and which you rather not want it to have.

## Features

**Open-Source:** This means you (or any developer you trust) can look up how things work inside the bot, can modify the bot yourself and be sure the bot does not use the access you are granting it to do stuff you don't want it to do. Also you are not dependent on some company to keep the bot running for you but are in control over it. In case I'm no longer willing to develop the bot, it will not cease to exist but can be developed further by anyone.

[**Overlays:**]({{< ref "overlays/_index.md" >}}) The bot contains a web-server to host custom overlays which can be built like any website with **HTML and Javascript**. Some default overlays are included ready to use and for everything not available in the default distribution there is a helper library available to connect to the bot and work with events and bot state.

**YAML Configuration:** The whole configuration is stored in a single YAML file, not in any proprietary format. You can simply create a backup of that file and even if some mistake or broken server happens you simply put back the configuration file and all of your rules, auto-messages, API-keys are instantly back again. (What's not in the configuration file is the data the bot stores like counters, events and variables.)

**Common Database Formats:** All the data mentioned in the last point is stored in a common database format like **SQLite**, **MySQL** or **PostgreSQL**. With exception of the credentials all the data is stored in a plain format which means you can use well-known database tooling to create backups. This also means you can use custom tooling to do with the data what you want! (Though a warning for this point: The database schema is not guaranteed to be stable! While it's possible I do not recommend directly accessing the data in the database for other tools.)

**API-First Design:** The bot is built to have an API and an included documentation for this API. Most of its functionality is exposed through the API and you can easily build tooling against that API to make to bot do your bidding.

[**Raffle-Module:**]({{< ref "modules/raffle.md" >}}) You don't need any other tool to do giveaways, the bot contains a raffle management including entrant restrictions, random picks, luck modifiers, automated text posts and of course entrant management.
