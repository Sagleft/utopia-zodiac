# utopia-zodiac
An example of WEB 3.0 horoscope bot

## Configuring the bot

1. [Download a ready-made release](releases) or build a bot from source.
2. [Create API key](https://rapidapi.com/Alejandro99aru/api/horoscope-astrology).
3. Fill in the data in the settings file: `config.json`

About fields in the config:
* `apikey` - API key that you received from rapidapi;
* `channelID` - channel ID in Utopia;
* `timeVariant` - Horoscope query type by time: `today`, `yesterday`, `tomorrow`, `week`, `month`, `year`;
* `wordReplace` - is used to replace the words in the received answer;
* `utopia` - Utopia client connection settings.

How to get parameters to connect to the Utopia API can be found [in this documentation](https://udocs.gitbook.io/utopia-api/utopia-api/how-to-enable-api-access).

## Build from sources

```bash
git clone https://github.com/Sagleft/utopia-zodiac
go build
```

## Useful Links

Looking for examples of projects for Utopia API? [Check out this documentation](https://udocs.gitbook.io/utopia-api/examples-of-projects).
