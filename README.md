# forgoven

Check for items presents in offline coop mates' inventory/enderchest and notify on Discord/IFTTT if a share item in present (Hypixel Skyblock).

### Usage

```sh
USAGE:
    forgoven -k HYPIXEL_API_KEY [-w DISCORD_WEBHOOK_URL] USERNAME|MINECRAFT_UUID:SKYBLOCK_PROFILE:DISCORD_USER_ID|NOTIGO_KEY:ITEM... ...

FLAGS:
    -h, --help       Prints help information

OPTIONS:
  -d duration
        time between two checks (default 1m0s)
  -k string
        hypixel API key
  -w string
        webhook url used to notify users on discord
```

### Example

```sh
forgoven \
    -k c447489f-52fe-4231-872c-803d17902e96 \
    -w https://discordapp.com/api/webhooks/758919348577719200/Sqs49JcaEo6N4vqctbsfwl0E6Jr-0XxpFUy8JdFQjGKWrYE9oLHn4Dsf9mNplucj1436 \
    scotow:Papaya:OQea02as-aze11:Stonk \
    'lrdoz:Papaya:981058438928120342:Stonk:Aspect of the Dragon' \
    boinc:Papaya:122004182860102840:Stonk:Minion
```

The command above:
- notify the user *Scotow* with [notigo](https://github.com/scotow/notigo) if he is disconnected with a *Stonk* item in his inventory/enderchest;
- notify the user *lrdoz* on Discord if he is disconnected with a *Stonk* or an *Aspect of the Dragon* item in his inventory/enderchest;
- notify the user *boinc* on Discord if he is disconnected with a *Stonk* or any *Minion* in his inventory/enderchest.