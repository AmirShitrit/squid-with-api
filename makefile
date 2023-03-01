listkeys:
	./bbolt.sh keys $(DBFILE) proxies

getvalue:
	./bbolt.sh get $(DBFILE) proxies $(KEY)
