import subprocess

num_iterations = 99

for i in range(num_iterations):
    quality = num_iterations - i
    subprocess.run(["bin/jpegme", "-in", f"{quality + 1}-iteration.jpg", "-out", f"{quality}-iteration.jpg", "-quality", f'{quality}'])
