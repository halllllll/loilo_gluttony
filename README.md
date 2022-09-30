# loilo_gluttony

scraping `LoiloNote` data.

[LoiloNote](https://n.loilo.tv/en/)

As a just only ICT support person, I created this for my own daily-work, but the company I work for (not related to systems development) found it and decided to make this public.

In my town, `LoiloNote` was adopted as one of the tools used in the [GIGA School Program](https://www.japantimes.co.jp/2021/03/22/special-supplements/japans-giga-school-program-equips-students-digital-society/), which began operating in Japan in 2020.

I have determined that according to robots.txt, this method of scraping is not currently prohibited. However, when using this method, please use your own judgment and modify it accordingly to avoid burdening the server.

(For reference, the LoiloNote accounts used by my client, City Board of Education, are approximately 180)


build ex:
```
GOOS=windows GOARCH=386 go build main.go
```