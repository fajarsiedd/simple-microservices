import { Test, TestingModule } from '@nestjs/testing';
import { ProductsService } from '../services/products.service';
import { RedisService } from '@liaoliaots/nestjs-redis';
import { IProductRepository } from '../interfaces/product-repository.interface';
import { RabbitmqService } from '../../shared/messaging/rabbitmq.service';
import { NotFoundException } from '@nestjs/common';

const mockProductRepository = () => ({
    create: jest.fn(),
    findAll: jest.fn(),
    findOneById: jest.fn(),
    update: jest.fn(),
    remove: jest.fn(),
});

const mockRabbitmqService = () => ({
    publishEvent: jest.fn(),
});

const mockRedisClient = {
    get: jest.fn(),
    set: jest.fn(),
    del: jest.fn(),
};

const mockRedisService = () => ({
    getOrThrow: jest.fn(() => mockRedisClient),
});

describe('ProductsService', () => {
    let service: ProductsService;
    let productRepository: ReturnType<typeof mockProductRepository>;
    let rabbitmqService: ReturnType<typeof mockRabbitmqService>;
    let redisClient: typeof mockRedisClient;
    let mockRedisServiceInstance: ReturnType<typeof mockRedisService>;

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            providers: [
                ProductsService,
                {
                    provide: 'IProductRepository',
                    useValue: mockProductRepository(),
                },
                {
                    provide: RabbitmqService,
                    useValue: mockRabbitmqService(),
                },
                {
                    provide: RedisService,
                    useValue: mockRedisService(),
                },
            ],
        }).compile();

        service = module.get<ProductsService>(ProductsService);
        productRepository = module.get<IProductRepository>('IProductRepository') as any;
        rabbitmqService = module.get<RabbitmqService>(RabbitmqService) as any;
        mockRedisServiceInstance = module.get<RedisService>(RedisService) as any;
        redisClient = mockRedisServiceInstance.getOrThrow();
    });

    it('should be defined', () => {
        expect(service).toBeDefined();
    });

    describe('reduceStock()', () => {
        it('should publish a "product.not-found" event if the product does not exist', async () => {
            productRepository.findOneById.mockResolvedValue(null);
            await service.reduceStock(1, 101, 5);
            expect(rabbitmqService.publishEvent).toHaveBeenCalledWith('product.not-found', expect.any(Object));
            expect(productRepository.update).not.toHaveBeenCalled();
            expect(redisClient.del).not.toHaveBeenCalled();
        });

        it('should publish a "stock-insufficient" event if there is not enough stock', async () => {
            const productWithLowStock = { id: 101, qty: 3 };
            productRepository.findOneById.mockResolvedValue(productWithLowStock);
            await service.reduceStock(1, 101, 5);
            expect(rabbitmqService.publishEvent).toHaveBeenCalledWith('product.stock-insufficient', expect.any(Object));
            expect(productRepository.update).not.toHaveBeenCalled();
            expect(redisClient.del).not.toHaveBeenCalled();
        });

        it('should update stock and clear cache if stock is sufficient', async () => {
            const productWithEnoughStock = { id: 101, qty: 10 };
            productRepository.findOneById.mockResolvedValue(productWithEnoughStock);
            productRepository.update.mockResolvedValue({ id: 101, qty: 5 });
            
            await service.reduceStock(1, 101, 5);
            
            expect(productRepository.update).toHaveBeenCalledWith(101, { qty: 5 });
            expect(redisClient.del).toHaveBeenCalledWith('products:id:101');
            expect(rabbitmqService.publishEvent).not.toHaveBeenCalled();
        });
    });

    describe('findOne()', () => {
        it('should return a product from cache if it exists', async () => {
            const product = { id: 1, name: 'Basmut', price: 1000 };
            redisClient.get.mockResolvedValue(JSON.stringify(product));
            
            const result = await service.findOne(1);
            
            expect(redisClient.get).toHaveBeenCalledWith('products:id:1');
            expect(productRepository.findOneById).not.toHaveBeenCalled();
            expect(result).toEqual(product);
        });

        it('should return a product from the repository and cache it if not in cache', async () => {
            const product = { id: 2, name: 'Buryam', price: 25 };
            redisClient.get.mockResolvedValue(null);
            productRepository.findOneById.mockResolvedValue(product);
            redisClient.set.mockResolvedValue('OK');
            
            const result = await service.findOne(2);

            expect(redisClient.get).toHaveBeenCalledWith('products:id:2');
            expect(productRepository.findOneById).toHaveBeenCalledWith(2);
            expect(redisClient.set).toHaveBeenCalledWith('products:id:2', JSON.stringify(product));
            expect(result).toEqual(product);
        });

        it('should throw NotFoundException if product does not exist', async () => {
            redisClient.get.mockResolvedValue(null);
            productRepository.findOneById.mockResolvedValue(null);

            await expect(service.findOne(999)).rejects.toThrow(NotFoundException);
        });
    });

    describe('create()', () => {
        it('should create a new product, clear cache, and publish an event', async () => {
            const createDto = { name: 'Nasgor', price: 20000 };
            const newProduct = { id: 1, ...createDto };
            
            productRepository.create.mockResolvedValue(newProduct);
            redisClient.del.mockResolvedValue(1);

            const result = await service.create(createDto as any);

            expect(productRepository.create).toHaveBeenCalledWith(createDto);
            expect(redisClient.del).toHaveBeenCalledWith('products:id:1');
            expect(rabbitmqService.publishEvent).toHaveBeenCalledWith('product.created', newProduct);
            expect(result).toEqual(newProduct);
        });
    });
});