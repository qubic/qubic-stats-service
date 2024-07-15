cd api || exit
sh buildDocker.sh
cd ..

cd processor || exit
sh buildDocker.sh
cd ..
