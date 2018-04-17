# Gomhotep

Antivirus on-access scanning for Linux using [ClamAV](https://www.clamav.net/) and [Fanotify](http://manpages.ubuntu.com/manpages/xenial/man7/fanotify.7.html)


### Dependencies:

Gomhotep depends on Go 1.6 and ClamAV to run. It's also important to install [freshclam](https://linux.die.net/man/1/freshclam) so ClamAV signatures are kept up to date.
On Ubuntu (16.04 LTS) install it with:

```
sudo apt-get install clamav libclamav-dev clamav-freshclam golang
```

### Configuration:

**1)** Edit `config/gomhotep.yml`.

Fanotify notifies events on a mounted filesystem so we need to provide a mountpoint to it. Currently Gomhotep supports only a single mount point (per gomhotep process).

Therefore, for `mount_point` use an existing mountpoint (like `/`) or create a temporary bind mount:

```
mkdir /tmp/gomhotep_base /tmp/gomhotep
sudo mount --bind /tmp/gomhotep_base /tmp/gomhotep/
```
and then update `mount_point` to `/tmp/gomhotep/`

**2)** Copy `config/gomhotep.yml` to `/etc/gomhotep/`


### Building / Running:

```
go build gomhotep.go
sudo ./gomhotep
```

Gomhotep will start the ClamAV scanning workers (defaults to 3 from `num_routines` on `config/gomhotep.yml`)  and load ClamAV's signature database on each.

After a couple of seconds it will display its status:

```
[0] initializing ClamAV database...
[1] initializing ClamAV database...
[2] initializing ClamAV database...
loaded 6471891 signatures
loaded 6471891 signatures
loaded 6471891 signatures
```

As soon as signatures are loaded it's ready to start scanning!

### Testing:

Download the [EICAR Anti-Virus Test File](http://www.eicar.org/download/eicar.com.txt) and place it anywhere on the chosen `mount_point`.

A `malware found` message should be displayed:
![Alt Text](https://s3.amazonaws.com/acmarques-github/ezgif-4-a28d1a34bf.gif)




##### Disclaimer:
*Gomhotep is a personal research project on filesystem event monitoring and not intended for production use*