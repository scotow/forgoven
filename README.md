# forgoven

### Features

- Check for items in offline coop mates' inventory/enderchest and notify them on Discord/IFTTT if a shared item is present (Hypixel Skyblock);
- Set Discord Channel's topic with online players;
- Send a message in a Discord channel with Oringo's today pets.

### Requirements

- Players need to allow access to their Inventory in their Skyblock settings;
- At least one Hypixel API key (generatable using the in-game command `/api`);
- (A Discord server);
- (A IFTTT API key).

### Usage

```sh
Usage:
  forgoven -k HYPIXEL_API_KEY... [-d CHECK_INTERVAL] [-w DISCORD_WEBHOOK_URL [-z]] [-t DISCORD_BOT_TOKEN -c DISCORD_CHANNEL_ID] -u USERNAME|MINECRAFT_UUID:SKYBLOCK_PROFILE:DISCORD_USER_ID|NOTIGO_KEY:ITEM... ...

Application Options:
  -k=         Hypixel API key(s)
  -d=         Time between two checks (default: 1m)
  -w=         Webhook url used to notify users on discord
  -u=         USERNAME|MINECRAFT_UUID:SKYBLOCK_PROFILE:DISCORD_USER_ID|NOTIGO_KEY:ITEM...
  -t=         Discord token used to update channel topic with online players
  -c=         Discord channel id used to update channel topic with online players
  -z          Check for Zoo pet and send it on Discord

Help Options:
  -h, --help  Show this help message
```

### Example

```sh
forgoven \
    -d 8s -z
    -k c447489f-52fe-4231-872c-803d17902e96 \
    -w https://discordapp.com/api/webhooks/758919348577719200/Sqs49JcaEo6N4vqctbsfwl0E6Jr-0XxpFUy8JdFQjGKWrYE9oLHn4Dsf9mNplucj1436 \
    -t AAAAAAAAAA.AAAAAA.AAAAAAAA -c 123456789012345678
    -u scotow:Papaya:OQea02as-aze11:Stonk \
    -u 'lrdoz:Papaya:981058438928120342:Stonk:Aspect of the Dragon' \
    -u boinc:Papaya:122004182860102840:Stonk:Minion
```

The command above:
- check inventory every 8s (3 players - 8s - 3 calls ~= 68 calls per minute, limit is 120);
- send Oringo's today pets in the channel;
- notify the user *Scotow* with [notigo](https://github.com/scotow/notigo) if he is disconnected with a *Stonk* item in his inventory/enderchest;
- notify the user *lrdoz* on Discord if he is disconnected with a *Stonk* or an *Aspect of the Dragon* item in his inventory/enderchest;
- notify the user *boinc* on Discord if he is disconnected with a *Stonk* or any *Minion* in his inventory/enderchest;
- update channel topic (id 123456789012345678) with the list of online players.
