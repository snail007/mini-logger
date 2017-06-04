# mini-logger
mini but flexible and powerful logger for go
# Notic
1.Do not call runtime.Goexit() in main , it will be blocking logger.Flush().   