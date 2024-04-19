#!/bin/bash

file_path="/etc/ssh/sshd_config"

docker_kill() {
    local isActive=$1
    if [ $isActive != "active" ]; then
        sudo pkill dockerd
        sleep 1s
        sudo systemctl restart docker   
        isActive=$(systemctl status docker | grep Active | awk {'print $2'})
        docker_kill "$isActive" 
    fi
}

check_SSHD() {
    local isActive=$1
    if [ $isActive == "active" ]; theni
        return 0
        else
            if [ -e "$file_path" ]; then
                # Extract the value of Port using grep
                port_value=$(grep "^Port" "$file_path" | awk '{print $2}')

                # replace the port value with 1022 or 22
                if [ "$port_value" == "22" ]; then
                                sudo sed -i '/Port/d' "$file_path"
                                sudo sed -i "15i\\Port 1022"   "$file_path"
                                sudo systemctl restart sshd

                                isActiveSSHD=$(systemctl status sshd | grep Active | awk {'print $2'})
                                check_SSHD "$isActiveSSHD" 

                else
                                sudo sed -i '/Port/d'   "$file_path"
                                sudo sed -i "15i\\Port 22"   "$file_path"
                                sudo systemctl restart sshd 

                                isActiveSSHD=$(systemctl status sshd | grep Active | awk {'print $2'})
                                check_SSHD "$isActiveSSHD" 
                fi
                else
                echo "File $file_path does not exist."
                return 1
            fi
    fi
}

isActive=$(systemctl status docker | grep Active | awk {'print $2'})
isActiveSSHD=$(systemctl status sshd | grep Active | awk {'print $2'})

# check_port func extracts the ssh port number and changes to either 1022 or 22
check_SSHD "$isActiveSSHD"

# docker_kill func kills the dockerd if systemd process is NOT active
docker_kill "$isActive"