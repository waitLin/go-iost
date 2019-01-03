class Exchange {
    init() {
    }

    /**
     *
     * @param tokenSym  {string}
     * @param from      {string}
     * @param to        {string}
     * @param amount    {string}
     * @param memo      {string}
     *
     * transfer("iost", "user0", "user1", "100.1", "")
     * transfer("iost", "user0", "", "100.1", "newUser2:OWNERKEY:ACTIVEKEY")
     */
    transfer(tokenSym, from, to, amount, memo) {
        const minAmount = 100;
        const initialRAM = 1000;
        const initialGasPledged = 10;

        let bamount = new BigNumber(amount);
        if (bamount.lt(minAmount)) {
            throw new Error("transfer amount should be greater or equal to " + minAmount);
        }

        if (from === "") {
            from = blockchain.publisher();
        }
        if (to !== "") {
            // transfer to an exist account
            blockchain.call("token.iost", "transfer", JSON.stringify([tokenSym, from, to, amount, memo]));

        } else if (to == "") {
            if (tokenSym !== "iost") {
                throw new Error("must transfer iost if you want to create a new account");
            }
            // create account and then transfer to account
            let args = memo.split(":");
            if (args.length !== 3) {
                throw new Error("memo of transferring to a new account should be of format name:ownerKey:activeKey");
            }
            // Attention: SignUp will use publisher() asset, so publisher should be equal to from
            blockchain.call("auth.iost", "SignUp", JSON.stringify(args));
            let rets = blockchain.call("ram.iost", "buy", JSON.stringify([from, args[0], initialRAM]));
            let price = rets[0];

            let paid = new BigNumber(price).plus(new BigNumber(initialGasPledged));
            if (bamount.lt(paid)) {
                throw new Error("amount not enough to buy 1kB RAM and pledge 10 IOST Gas. need " + bamount.toString())
            }

            blockchain.transfer(from, args[0], bamount.minus(paid), "");
        }
    }

}

module.exports = Exchange;
