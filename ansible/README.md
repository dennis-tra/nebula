# Nebula Crawler Deployment

## Configuration

### `ansible.cfg`

Add a line to the `ansible.cfg` file with IP-Address or host name of your VM under `hosts`:

```ini
[hosts]
123.456.123.456 ansible_connection=ssh ansible_user=nebula ansible_python_interpreter=/usr/bin/python3
```

### `ansible.vault`

Take the `ansible.vault.example` file and make a copy that is named `ansible.vault`.
The configuration parameters have comments close to them.

By editing items under the `networks` key you can add and remove networks that will be crawled and monitored.

## Preparing Host VM

If you rented a VM and have root access then run this command first:

```shell
ansible-playbook -i ansible.cfg --user root root.yml
```

This script will change the password of the `root` user, add new user to the `sudo` group and enable public key authentication.

## Setting Up Host VM

```shell
ansible-playbook --ask-sudo-pass -i ansible.cfg setup.yml
```

In China:

`/etc/docker/daemon.json`:

```json
{
  "registry-mirrors": [
    "https://registry.docker-cn.com"
  ]
}
```
and run instead:

```shell
ansible-playbook --ask-sudo-pass -i ansible.cfg setup.yml --extra-vars "use_docker_mirror=true"
```
