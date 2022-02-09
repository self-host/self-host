all: aapije juvuln malgomaj

aapije:
	go build github.com/self-host/self-host/cmd/aapije

juvuln:
	go build github.com/self-host/self-host/cmd/juvuln

malgomaj:
	go build github.com/self-host/self-host/cmd/malgomaj

clean:
	rm aapije malgomaj juvuln
