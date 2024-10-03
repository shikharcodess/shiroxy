current = 5
length = 10

index = (current + 0) % length
# print(f"({index} + 0) % {length} = {index}")

for i in range(length-1):
    # if (i == 0):
    #     print(f"${i}. ({index} + 1) % {length} = {index}")

    index = (index + 1) % length
    print(f"${i}. ({index} + 1) % {length} = {index}")
