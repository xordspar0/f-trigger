# 𝑓-trigger

Trigger events based on a frequency on stdin

## Examples

```sh
cycling_cadence.py | f-trigger \
    0 "play brake_song.ogg" \
    1-80 "play easy_song.ogg" \
    81- "play pumping_song.ogg"
```

```sh
tail -f /var/log/apache2/access.log | f-trigger 10000- "play alert.ogg"
```

<p align="center" ><img src="https://user-images.githubusercontent.com/2390950/35239547-67bfc0a4-ff76-11e7-90b9-244ec816db3f.png" /></p>
