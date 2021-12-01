# Nebula Crawler Deployment

## Configuration

### `ansible.cfg`

Add a line to the `ansible.cfg` file with IP-Address or host name of your VM under `hosts`:

```ini
[hosts]
# hosts where you have SSH access to a user in the sudo group (the scripts expect the username to be nebula)
123.456.123.456 ansible_connection=ssh ansible_user=nebula ansible_python_interpreter=/usr/bin/python3

[root]
# hosts where you have SSH access to the root user
123.456.123.456 ansible_connection=ssh ansible_user=nebula ansible_python_interpreter=/usr/bin/python3
```

### `ansible.vault`

Take the `ansible.vault.example` file and make a copy that is named `ansible.vault`.
The configuration parameters have comments close to them.

By editing items under the `networks` key you can add and remove networks that will be crawled and monitored.

## `root.yml` - Preparing Host VM 

If you rented a VM and have root access then run this command first:

```shell
ansible-playbook -i ansible.cfg root.yml
```

This script will change the password of the `root` user, add new user to the `sudo` group and enable public key authentication.

## `setup.yml` - Setting Up Host VM

This script assumes that on the remote system exists a user name `{{ remote_user_name }}` (see `ansible.vault.example`) that has sudo rights.

```shell
ansible-playbook -i ansible.cfg setup.yml
```

If your node is in China set `use_docker_mirror` to true in your `ansible.cfg`. This will add the following JSON to `/etc/docker/daemon.json`:

```json
{
  "registry-mirrors": [
    "https://registry.docker-cn.com"
  ]
}
```

## `deploy.yml` - Deploying the Docker  Containers

```shell
ansible-playbook -i ansible.cfg deploy.yml
```
