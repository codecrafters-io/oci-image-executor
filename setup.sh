set -e

echo "installing docker"
sudo apt-get update
sudo apt-get install -y ca-certificates curl gnupg lsb-release
curl -fsSL https://get.docker.com -o get-docker.sh
sh ./get-docker.sh

echo "installing firecracker"
curl -SL https://github.com/firecracker-microvm/firecracker/releases/download/v0.25.2/firecracker-v0.25.2-x86_64.tgz -o firecracker.tgz
tar -xf firecracker.tgz
sudo cp release-v0.25.2-x86_64/firecracker-v0.25.2-x86_64 /usr/bin/firecracker
rm -rf release-v0.25.2-x86_64 firecracker.tgz

echo "installing make"
sudo apt-get install -y make

sudo apt-get install -y cpu-checker # adds support for kvm-ok binary

echo "installing go"
wget -c https://dl.google.com/go/go1.14.2.linux-amd64.tar.gz -O - | sudo tar -xz -C /usr/local
echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.bashrc
export PATH=$PATH:/usr/local/go/bin

mkdir -p /etc/cni/conf.d
cp fcnet.conflist /etc/cni/conf.d/

mkdir -p /opt/cni/bin

rm -rf plugins
git clone https://github.com/containernetworking/plugins
pushd plugins
git checkout v0.8.7
./build_linux.sh
cp bin/ptp /opt/cni/bin
cp bin/host-local /opt/cni/bin
cp bin/firewall /opt/cni/bin
popd

rm -rf tc-redirect-tap
git clone https://github.com/awslabs/tc-redirect-tap
pushd tc-redirect-tap
git checkout a0300978797dabc3b4ffaa4a30817d6e8dd10018
make
cp tc-redirect-tap /opt/cni/bin
popd

mkdir -p /root/firecracker-resources
curl -sS --fail -Lo /root/firecracker-resources/vmlinux.bin https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin
