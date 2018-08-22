# redis 分布式锁实现

## 加锁实现

SETNX 可以直接加锁操作，比如说对某个关键词foo加锁，客户端可以尝试`SETNX foo.lock <current unix time>`;
如果返回1，表示客户端已经获取锁，可以往下操作，操作完成后，通过`DEL foo.lock`命令来释放锁。
仅仅使用SETNX加锁带有竞争条件的，在某些特定的情况会造成死锁错误。

__考虑几种情况：__


1. setnx

```
C1获取锁，并崩溃。C2和C3调用SETNX上锁返回0后，获得foo.lock的时间戳，通过比对时间戳，发现锁超时。
C2 向foo.lock发送DEL命令。
C2 向foo.lock发送SETNX获取锁。
C3 向foo.lock发送DEL命令，此时C3发送DEL时，其实DEL掉的是C2的锁。
C3 向foo.lock发送SETNX获取锁。
```

2. getset

```
C1获取锁，并崩溃。C2和C3调用SETNX上锁返回0后，调用GET命令获得foo.lock的时间戳T1，通过比对时间戳，发现锁超时。
C4 向foo.lock发送GESET命令，
GETSET foo.lock <current unix time>
并得到foo.lock中老的时间戳T2

如果T1=T2，说明C4获得时间戳。
如果T1!=T2，说明C4之前有另外一个客户端C5通过调用GETSET方式获取了时间戳，C4未获得锁。只能sleep下，进入下次循环中。
```

3.1 __GET返回nil时应该走那种逻辑？__

```
第一种走超时逻辑
C1客户端获取锁，并且处理完后，DEL掉锁，在DEL锁之前。C2通过SETNX向foo.lock设置时间戳T0 发现有客户端获取锁，进入GET操作。
C2 向foo.lock发送GET命令，获取返回值T1(nil)。
C2 通过T0>T1+expire对比，进入GETSET流程。
C2 调用GETSET向foo.lock发送T0时间戳，返回foo.lock的原值T2
C2 如果T2=T1相等，获得锁，如果T2!=T1，未获得锁。
```

3.2 __GET返回nil时应该走那种逻辑？__

```
第二种情况走循环走setnx逻辑
C1客户端获取锁，并且处理完后，DEL掉锁，在DEL锁之前。C2通过SETNX向foo.lock设置时间戳T0 发现有客户端获取锁，进入GET操作。
C2 向foo.lock发送GET命令，获取返回值T1(nil)。
C2 循环，进入下一次SETNX逻辑
```

两种逻辑貌似都是OK，但是从逻辑处理上来说，第一种情况存在问题。当GET返回nil表示，锁是被删除的，而不是超时，应该走SETNX逻辑加锁。走第一种情况的问题是，正常的加锁逻辑应该走SETNX，而现在当锁被解除后，走的是GETST，如果判断条件不当，就会引起死锁，具体怎么碰到的看下面的问题:

__GETSET返回nil时应该怎么处理？__

```
C1和C2客户端调用GET接口，C1返回T1，此时C3网络情况更好，快速进入获取锁，并执行DEL删除锁，C2返回T2(nil)，C1和C2都进入超时处理逻辑。
C1 向foo.lock发送GETSET命令，获取返回值T11(nil)。
C1 比对C1和C11发现两者不同，处理逻辑认为未获取锁。
C2 向foo.lock发送GETSET命令，获取返回值T22(C1写入的时间戳)。
C2 比对C2和C22发现两者不同，处理逻辑认为未获取锁。
```

此时C1和C2都认为未获取锁，其实C1是已经获取锁了，但是他的处理逻辑没有考虑GETSET返回nil的情况，只是单纯的用GET和GETSET值就行对比，至于为什么会出现这种情况？一种是多客户端时，每个客户端连接Redis的后，发出的命令并不是连续的，导致从单客户端看到的好像连续的命令，到Redis server后，这两条命令之间可能已经插入大量的其他客户端发出的命令，比如DEL,SETNX等。第二种情况，多客户端之间时间不同步，或者不是严格意义的同步。
