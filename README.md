# Proxyscore

Command line interface to evaluate the anonymity of a given proxy.

## Anonymity

To understand the different level have a look at http://www.proxynova.com/proxy-articles/proxy-anonymity-levels-explained/

## Install

Either

* Install Go beforehand
* Run `got get github.com/simcap/proxyscore/cmd/proxyscore`

Or

* Get the executable (built for `darwin_386`) at https://github.com/simcap/proxyscore/releases/tag/v0.9


## Usage

At your command line

    $ proxyscore -p 164.215.111.16:80     // proxy given as host:port

... would give you the following json in stdout

    {
      "Anonymous": true,
      "Score": 1,
      "MyIP": "184.234.56.78",
      "Proxy": "164.215.111.16:80"
      "IPdetection": [  // show where the target can see your ip
        "RemoteAddr": "184.234.56.78",
        "X-Forwarded-For": "184.234.56.78"
      ],
      "Proxydetection": [ // show what the target sees as proxy info
        "Via": "1.1 10.234.128.2 (Mikrotik HttpProxy)",
        "X-Proxy-Id": "1648484578",
        "X-Forwarded-For": "184.234.56.78"
      ]
    }
