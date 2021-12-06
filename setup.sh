set -e

echo "installing docker"
sudo apt-get update
sudo apt-get install -y ca-certificates curl gnupg lsb-release
curl -fSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update
sudo apt-get install -y docker-ce=5:20.10.11~3-0~ubuntu-focal docker-ce-cli=5:20.10.11~3-0~ubuntu-focal containerd.io

echo "installing firecracker"
curl -SL https://github.com/firecracker-microvm/firecracker/releases/download/v0.25.2/firecracker-v0.25.2-x86_64.tgz -o firecracker.tgz
tar -xf firecracker.tgz
sudo cp release-v0.25.2-x86_64/firecracker-v0.25.2-x86_64 /usr/bin/firecracker
sudo cp release-v0.25.2-x86_64/jailer-v0.25.2-x86_64 /usr/bin/jailer
rm -rf release-v0.25.2-x86_64 firecracker.tgz

echo "installing make"
sudo apt-get install -y make

sudo apt-get install -y cpu-checker # adds support for kvm-ok binary

echo "installing go"
wget -c https://dl.google.com/go/go1.14.2.linux-amd64.tar.gz -O - | sudo tar -xz -C /usr/local
echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.bashrc

mkdir -p /etc/cni/conf.d
cp fcnet.conflist /etc/cni/conf.d/

mkdir /opt/cni/bin

git clone https://github.com/containernetworking/plugins
pushd plugins
git checkout v0.8.7
./build_linux.sh
cp bin/ptp /opt/cni/bin
cp bin/host-local /opt/cni/bin
cp bin/firewall /opt/cni/bin
popd

git clone https://github.com/awslabs/tc-redirect-tap
pushd tc-redirect-tap
make
cp tc-redirect-tap /opt/cni/bin
popd
