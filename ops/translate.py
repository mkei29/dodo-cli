import argparse

SYSTEM_PROMPT = """
You are about to translate a text from one language to another.
"""

def main():
    parser = argparse.ArgumentParser(description="Translate a text to another language")
    parser.add_argument("text", help="The text to translate")
    parser.add_argument("-i", "--input-language", help="The language to translate to", default="en")
    parser.add_argument("-o", "--output-language", help="The language to translate to", default="en")

    args = parser.parse_args()
    print(args)


if __name__ == "__main__":
    main()