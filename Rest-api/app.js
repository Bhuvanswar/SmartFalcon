import express, { json } from 'express';
import { Gateway, Wallets } from 'fabric-network';
import { resolve, join } from 'path';
import { readFileSync } from 'fs';

const app = express();
app.use(json());
const connectToFabric = async () => {
    const ccpPath = resolve('Reset-api', 'connection.json'); 
    const ccp = JSON.parse(readFileSync(ccpPath, 'utf8'));

    const walletPath = join(process.cwd(), 'wallet');  
    const wallet = await Wallets.newFileSystemWallet(walletPath);

    const gateway = new Gateway();
    await gateway.connect(ccp, {
        wallet,
        identity: 'User1', 
        discovery: { enabled: true, asLocalhost: true }  
    });

    const network = await gateway.getNetwork('mychannel');
    const contract = network.getContract('asset-transfer-basic'); 
    return contract;
};
app.post('/assets', async (req, res) => {
    try {
        const contract = await connectToFabric();
        const { id, value } = req.body;
        await contract.submitTransaction('CreateAsset', id, value);
        res.status(201).send({ message: 'Asset created successfully' });
    } catch (error) {
        res.status(500).send({ error: error.message });
    }
});
app.get('/assets/:id', async (req, res) => {
    try {
        const contract = await connectToFabric();
        const result = await contract.evaluateTransaction('ReadAsset', req.params.id); 
        res.status(200).send(result.toString());
    } catch (error) {
        res.status(500).send({ error: error.message });
    }
});
app.listen(3000, () => {
    console.log('API server running on port 3000');
});
