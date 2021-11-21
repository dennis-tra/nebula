# Nebula Crawler Deployment

ansible-playbook -i ansible.cfg --user root setup.yml
ansible-playbook -i ansible.cfg --user nebula deploy.yml
ansible-playbook -i ansible.cfg --user nebula update.yml
